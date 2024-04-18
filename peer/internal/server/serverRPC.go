/*
 *	References:
 *		https://gist.github.com/upperwal/38cd0c98e4a6b34c061db0ff26def9b9
 *		https://ldej.nl/post/building-an-echo-application-with-libp2p/
 *		https://github.com/libp2p/go-libp2p/blob/master/examples/chat-with-rendezvous/chat.go
 *		https://github.com/libp2p/go-libp2p/blob/master/examples/pubsub/basic-chat-with-rendezvous/main.go
 */

package server

import (
	"bufio"
	"context"
	"crypto/rsa"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"orca-peer/internal/fileshare"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	record "github.com/libp2p/go-libp2p-record"
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type fileShareServerNode struct {
	fileshare.UnimplementedFileShareServer
	savedFiles   map[string][]*fileshare.FileDesc // read-only after initialized
	mu           sync.Mutex                       // protects routeNotes
	currentCoins float64

	K_DHT   *dht.IpfsDHT
	PrivKey libp2pcrypto.PrivKey
	PubKey  libp2pcrypto.PubKey
	V       record.Validator
}

var (
	serverStruct fileShareServerNode
)

func CreateMarketServer(stdPrivKey *rsa.PrivateKey, dhtPort string, rpcPort string, serverReady chan bool) {
	ctx := context.Background()

	//Get libp2p wrapped privKey
	privKey, _, err := libp2pcrypto.KeyPairFromStdKey(stdPrivKey)
	if err != nil {
		panic("Could not generate libp2p wrapped key from standard private key.")
	}

	pubKey := privKey.GetPublic()

	//Construct multiaddr from string and create host to listen on it
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", dhtPort))
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(sourceMultiAddr.String()),
		libp2p.Identity(privKey), //derive id from private key
	}

	host, err := libp2p.New(opts...)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nlibp2p DHT Host ID: %s\n", host.ID())
	fmt.Println("DHT Market Multiaddr (if in server mode):")
	for _, addr := range host.Addrs() {
		fmt.Printf("%s/p2p/%s\n", addr, host.ID())
	}

	bootstrapPeers := ReadBootstrapPeers()

	// Start a DHT, for now we will start in client mode until we can implement a way to
	// detect if we are behind a NAT or not to run in server mode.
	var validator record.Validator = OrcaValidator{}
	var options []dht.Option
	options = append(options, dht.Mode(dht.ModeClient))
	options = append(options, dht.ProtocolPrefix("orcanet/market"), dht.Validator(validator))
	kDHT, err := dht.New(ctx, host, options...)
	if err != nil {
		panic(err)
	}

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	if err = kDHT.Bootstrap(ctx); err != nil {
		panic(err)
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	var wg sync.WaitGroup
	for _, peerAddr := range bootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := host.Connect(ctx, *peerinfo); err != nil {
				fmt.Println("WARNING: ", err)
			} else {
				fmt.Println("Connection established with DHT bootstrap node:", *peerinfo)
			}
		}()
	}
	wg.Wait()

	go DiscoverPeers(ctx, host, kDHT, "orcanet/market")

	//Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", rpcPort))
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	serverStruct = fileShareServerNode{}
	serverStruct.K_DHT = kDHT
	serverStruct.PrivKey = privKey
	serverStruct.PubKey = pubKey
	serverStruct.V = validator
	fileshare.RegisterFileShareServer(s, &serverStruct)
	go ListAllDHTPeers(ctx, host)
	fmt.Printf("Market RPC Server listening at %v\n\n", lis.Addr())
	serverReady <- true
	if err := s.Serve(lis); err != nil {
		panic(err)
	}
}
func ListAllDHTPeers(ctx context.Context, host host.Host) {
	for {
		time.Sleep(time.Second * 10)
		peers := serverStruct.K_DHT.RoutingTable().ListPeers()
		// Should make a channel that waits for this

		for _, p := range peers {
			_, err := serverStruct.K_DHT.FindPeer(ctx, p)
			if err != nil {
				fmt.Printf("Error finding peer %s: %s\n", p, err)
				continue
			}
		}
	}

}

/*
 * Check for peers who have announced themselves on the DHT.
 * If the DHT is running in server mode, then we will announce ourselves and check for
 * others who have announced as well.
 *
 * Parameters:
 *   context: The context
 *   h: libp2p host
 *   kDHT: the libp2p ipfs DHT object to use
 *   advertise: the string to use to check for others who have announced themselves. If
 * 				DHT is in server mode then that string will be used to announce ourselves as well.
 *
 */
func DiscoverPeers(ctx context.Context, h host.Host, kDHT *dht.IpfsDHT, advertise string) {
	routingDiscovery := drouting.NewRoutingDiscovery(kDHT)
	if kDHT.Mode() == dht.ModeServer {
		dutil.Advertise(ctx, routingDiscovery, advertise)
	}

	// Look for others who have announced and attempt to connect to them
	for {
		peerChan, err := routingDiscovery.FindPeers(ctx, advertise)
		if err != nil {
			panic(err)
		}
		for peer := range peerChan {
			if peer.ID == h.ID() {
				continue // No self connection
			}
			h.Connect(ctx, peer)
		}
		time.Sleep(time.Second * 10)
	}
}

func sendFileToConsumer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		for k, v := range r.URL.Query() {
			fmt.Printf("%s: %s\n", k, v)
		}
		// file = r.URL.Query().Get("filename")
		w.Write([]byte("Received a GET request\n"))

	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
	w.Write([]byte("Received a GET request\n"))
	filename := r.URL.Path[len("/reqFile/"):]

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer file.Close()

	// Set content type
	contentType := "application/octet-stream"
	switch {
	case filename[len(filename)-4:] == ".txt":
		contentType = "text/plain"
	case filename[len(filename)-5:] == ".json":
		contentType = "application/json"
		// Add more cases for other file types if needed
	}

	// Set content disposition header
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", contentType)

	// Copy file contents to response body
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func runNotifyStore(client fileshare.FileShareClient, file *fileshare.FileDesc) *fileshare.StorageACKResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ackResponse, err := client.NotifyFileStore(ctx, file)
	if err != nil {
		log.Fatalf("client.NotifyFileStorage failed: %v", err)
	}
	log.Printf("ACK Response: %v", ackResponse)
	return ackResponse
}

func runNotifyUnstore(client fileshare.FileShareClient, file *fileshare.FileDesc) *fileshare.StorageACKResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ackResponse, err := client.NotifyFileUnstore(ctx, file)
	if err != nil {
		log.Fatalf("client.NotifyFileStorage failed: %v", err)
	}
	log.Printf("ACK Response: %v", ackResponse)
	return ackResponse
}

func NotifyStoreWrapper(client fileshare.FileShareClient, file_name_hash string, file_name string, file_size_bytes int64, file_origin_address string, origin_user_id string, file_cost float32, file_data_hash string, file_bytes []byte) {
	var file_description = fileshare.FileDesc{FileNameHash: file_name_hash,
		FileName:          file_name,
		FileSizeBytes:     file_size_bytes,
		FileOriginAddress: file_origin_address,
		OriginUserId:      origin_user_id,
		FileCost:          file_cost,
		FileDataHash:      file_data_hash,
		FileBytes:         file_bytes}
	var ack = runNotifyUnstore(client, &file_description)
	if ack.IsAcknowledged {
		fmt.Printf("[Server]: Market acknowledged stopping storage of file %s with hash %s \n", ack.FileName, ack.FileHash)
	} else {
		fmt.Printf("[Server]: Unable to notify market that we are stopping the storage of file %s with hash %s \n", ack.FileName, ack.FileHash)
	}
}
func NotifyUnstoreWrapper(client fileshare.FileShareClient, file_name_hash string, file_name string, file_size_bytes int64, file_origin_address string, origin_user_id string, file_cost float32, file_data_hash string, file_bytes []byte) {
	var file_description = fileshare.FileDesc{FileNameHash: file_name_hash,
		FileName:          file_name,
		FileSizeBytes:     file_size_bytes,
		FileOriginAddress: file_origin_address,
		OriginUserId:      origin_user_id,
		FileCost:          file_cost,
		FileDataHash:      file_data_hash,
		FileBytes:         file_bytes}
	var ack = runNotifyUnstore(client, &file_description)
	if ack.IsAcknowledged {
		fmt.Printf("[Server]: Market acknowledged stopping storage of file %s with hash %s \n", ack.FileName, ack.FileHash)
	} else {
		fmt.Printf("[Server]: Unable to notify market that we are stopping the storage of file %s with hash %s \n", ack.FileName, ack.FileHash)
	}
}

func SetupRegisterFile(fileHash string, amountPerMB int64, ip string, port int32) error {
	ctx := context.Background()
	fileReq := fileshare.RegisterFileRequest{}
	fileReq.User = &fileshare.User{}
	fileReq.User.Price = amountPerMB
	fileReq.User.Ip = ip
	fileReq.User.Port = port
	fileReq.FileHash = fileHash
	_, err := serverStruct.RegisterFile(ctx, &fileReq)
	if err != nil {
		return err
	}
	return nil
}

/*
 * gRPC service to register a file on the DHT market.
 *
 * Parameters:
 *   ctx: Context
 *   in: A protobuf RegisterFileRequest struct that represents the file/producer being registered.
 *
 * Returns:
 *   An empty protobuf struct
 *   An error, if any
 */
func (s *fileShareServerNode) RegisterFile(ctx context.Context, in *fileshare.RegisterFileRequest) (*emptypb.Empty, error) {
	hash := in.GetFileHash()
	pubKeyBytes, err := s.PubKey.Raw()
	if err != nil {
		return nil, err
	}
	in.GetUser().Id = pubKeyBytes

	value, err := s.K_DHT.GetValue(ctx, "orcanet/market/"+hash)
	if err != nil {
		value = make([]byte, 0)
	}

	//remove record for id if it already exists
	for i := 0; i < len(value)-8; i++ {
		messageLength := uint16(value[i+1])<<8 | uint16(value[i])
		digitalSignatureLength := uint16(value[i+3])<<8 | uint16(value[i+2])
		contentLength := messageLength + digitalSignatureLength
		user := &fileshare.User{}

		err := proto.Unmarshal(value[i+4:i+4+int(messageLength)], user) //will parse bytes only until user struct is filled out
		if err != nil {
			return nil, err
		}

		if len(user.GetId()) == len(in.GetUser().GetId()) {
			recordExists := true
			for i := range in.GetUser().GetId() {
				if user.GetId()[i] != in.GetUser().GetId()[i] {
					recordExists = false
					break
				}
			}

			if recordExists {
				value = append(value[:i], value[i+4+int(contentLength):]...)
				break
			}
		}

		i = i + 4 + int(contentLength) - 1
	}

	record := make([]byte, 0)
	userProtoBytes, err := proto.Marshal(in.GetUser())
	if err != nil {
		return nil, err
	}
	userProtoSize := len(userProtoBytes)
	signature, err := s.PrivKey.Sign(userProtoBytes)
	if err != nil {
		return nil, err
	}
	signatureLength := len(signature)
	record = append(record, byte(userProtoSize))
	record = append(record, byte(userProtoSize>>8))
	record = append(record, byte(signatureLength))
	record = append(record, byte(signatureLength>>8))
	record = append(record, userProtoBytes...)
	record = append(record, signature...)

	currentTime := time.Now().UTC()
	unixTimestamp := currentTime.Unix()
	unixTimestampInt64 := uint64(unixTimestamp)
	for i := 7; i >= 0; i-- {
		curByte := unixTimestampInt64 >> (i * 8)
		record = append(record, byte(curByte))
	}

	if len(value) != 0 {
		value = value[:len(value)-8] //get rid of previous values timestamp
	}
	value = append(value, record...)

	err = s.K_DHT.PutValue(ctx, "orcanet/market/"+in.GetFileHash(), value)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func SetupCheckHolders(fileHash string) (*fileshare.HoldersResponse, error) {
	ctx := context.Background()
	fileReq := fileshare.CheckHoldersRequest{}
	fileReq.FileHash = fileHash
	holdersResponse, err := serverStruct.CheckHolders(ctx, &fileReq)
	if err != nil {
		return nil, err
	}
	return holdersResponse, nil
}

/*
 * gRPC service to check for producers who have registered a specific file.
 *
 * Parameters:
 *   ctx: Context
 *   in: A protobuf CheckHoldersRequest struct that represents the file to look up.
 *
 * Returns:
 *   A HoldersResponse protobuf struct that represents the producers and their prices.
 *   An error, if any
 */
func (s *fileShareServerNode) CheckHolders(ctx context.Context, in *fileshare.CheckHoldersRequest) (*fileshare.HoldersResponse, error) {
	hash := in.GetFileHash()
	users := make([]*fileshare.User, 0)
	value, err := s.K_DHT.GetValue(ctx, "orcanet/market/"+hash)
	if err != nil {
		return &fileshare.HoldersResponse{Holders: users}, nil
	}

	for i := 0; i < len(value)-8; i++ {
		messageLength := uint16(value[i+1])<<8 | uint16(value[i])
		digitalSignatureLength := uint16(value[i+3])<<8 | uint16(value[i+2])
		contentLength := messageLength + digitalSignatureLength
		user := &fileshare.User{}

		err := proto.Unmarshal(value[i+4:i+4+int(messageLength)], user) //will parse bytes only until user struct is filled out
		if err != nil {
			return nil, err
		}

		users = append(users, user)
		i = i + 4 + int(contentLength) - 1
	}

	return &fileshare.HoldersResponse{Holders: users}, nil
}

// Find file bootstrap.peers and parse it to get multiaddrs of bootstrap peers
func ReadBootstrapPeers() []multiaddr.Multiaddr {
	peers := []multiaddr.Multiaddr{}

	// For now bootstrap.peers can be in cli folder but it can be moved
	file, err := os.Open("internal/cli/bootstrap.peers")
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()

		multiadd, err := multiaddr.NewMultiaddr(line)
		if err != nil {
			panic(err)
		}
		peers = append(peers, multiadd)
	}

	return peers
}
