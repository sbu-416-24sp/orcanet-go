package server

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	orcaClient "orca-peer/internal/client"
	"orca-peer/internal/fileshare"
	"orca-peer/internal/hash"
	orcaJobs "orca-peer/internal/jobs"
	"github.com/libp2p/go-libp2p/core/host"
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"os"
	"path/filepath"
	"time"
	"github.com/google/uuid"
)

const keyServerAddr = "serverAddr"

var (
	eventChannel chan bool
	Client       *orcaClient.Client
	PassKey      string
)

type HTTPServer struct {
	storage *hash.DataStore
}

type TransactionFile struct {
	Bytes               []byte  `json:"bytes"`
	UnlockedTransaction []byte  `json:"transaction"`
	PublicKey           string  `json:"public_key"`
	Date                string  `json:"date"`
	Cost                float64 `json:"cost"`
}
type Transaction struct {
	Price     float64 `json:"price"`
	Timestamp string  `json:"timestamp"`
	Uuid      string  `json:"uuid"`
}

func handleTransaction(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Print the received byte string
	var data TransactionFile
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}
	publicKey, err := hash.ParseRsaPublicKeyFromPemStr(data.PublicKey)
	if err != nil {
		fmt.Println("Unable to unmarshalling public key:", err)
		return
	}
	timestamp := time.Now()
	timestampStr := timestamp.Format(time.RFC3339Nano)
	err = os.WriteFile("./files/transactions/"+timestampStr, body, 0644)
	if err != nil {
		fmt.Println("Error writing transaction to file:", err)
		return
	}
	error := hash.VerifySignature(data.UnlockedTransaction, data.Bytes, publicKey)
	if error != nil {
		fmt.Println("Properly Hashed Transaction")
	} else {
		fmt.Println("Did not properly hash transaction")
	}
	var transaction Transaction
	err = json.Unmarshal(data.UnlockedTransaction, &transaction)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	} else {
		fmt.Println("Transaction JSON:")
		fmt.Println("Sent Amount:")
		fmt.Println(transaction.Price)
		fmt.Println("UUID:")
		fmt.Println(transaction.Uuid)
	}
	eventChannel <- true
	fmt.Println("> ")
}

// Start HTTP/RPC server
func StartServer(httpPort string, dhtPort string, rpcPort string, serverReady chan bool, confirming *bool, confirmation *string, libp2pPrivKey libp2pcrypto.PrivKey, passKey string, client *orcaClient.Client, startAPIRoutes func(*map[string]fileshare.FileInfo), host host.Host) {
	eventChannel = make(chan bool)
	server := HTTPServer{
		storage: hash.NewDataStore("files/stored/"),
	}
	go orcaJobs.InitPeriodicJobSave()
	Client = client
	PassKey = passKey
	fileShareServer := FileShareServerNode{
		StoredFileInfoMap: make(map[string]fileshare.FileInfo),
	}

	//Why are there routes in 2 different spots?
	http.HandleFunc("/requestFile/", func(w http.ResponseWriter, r *http.Request) {
		server.sendFile(w, r, confirming, confirmation)
	})
	http.HandleFunc("/storeFile/", func(w http.ResponseWriter, r *http.Request) {
		server.storeFile(w, r, confirming, confirmation)
	})
	http.HandleFunc("/sendTransaction", handleTransaction)
	http.HandleFunc("/get-peers", getAllPeers)
	http.HandleFunc("/get-peer", getPeer)
	http.HandleFunc("/find-peer", FindPeersForHash)
	http.HandleFunc("/remove-peer", removePeer)

	http.HandleFunc("/add-job", AddJobHandler)

	fmt.Printf("HTTP Listening on port %s...\n", httpPort)
	go CreateMarketServer(libp2pPrivKey, dhtPort, rpcPort, serverReady, &fileShareServer, host)
	startAPIRoutes(&fileShareServer.StoredFileInfoMap)

	http.ListenAndServe(":"+httpPort, nil)
}

type Peer struct {
	PeerId string  `json:"peerID"`
	Ip     string  `json:"ip"`
	Region string  `json:"region"`
	Price  float32 `json:"price"`
}

func FindPeersForHash(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		queryParams := r.URL.Query()
		hash := queryParams.Get("fileHash")
		peers, err := findPeersForHash(hash)
		if err != nil {
			w.WriteHeader(http.StatusMethodNotAllowed)
			writeStatusUpdate(w, "Errors retrieving information about peers holding this hash.")
			return
		}
		jsonData, err := json.Marshal(peers)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeStatusUpdate(w, "Failed to convert JSON Data into a string")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only PUT requests will be handled.")
		return
	}
}

func findPeersForHash(fileHash string) ([]Peer, error) {
	holders, err := SetupCheckHolders(fileHash)
	if err != nil {
		return []Peer{}, err
	}
	peers := make([]Peer, 0)
	for _, holder := range holders.Holders {
		location, err := getLocationFromIP(string(holder.GetId()))
		if err != nil {
			return []Peer{}, errors.New("unable to get location about peer")
		}
		peers = append(peers, Peer{
			PeerId: string(holder.GetId()),
			Ip:     holder.Ip,
			Region: location,
		})
	}
	return peers, nil
}

func (server *HTTPServer) sendFile(w http.ResponseWriter, r *http.Request, confirming *bool, confirmation *string) {
	// Extract filename from URL path
	filename := r.URL.Path[len("/requestFile/"):]

	// Ask for confirmation
	// *confirming = true
	// fmt.Printf("You have just received a request to send file '%s'. Do you want to send the file? (yes/no): ", filename)

	// // Check if confirmation is received
	// for *confirmation != "yes" {
	// 	if *confirmation != "" {
	// 		http.Error(w, fmt.Sprintf("Client declined to send file '%s'.", filename), http.StatusUnauthorized)
	// 		*confirmation = ""
	// 		*confirming = false
	// 		return
	// 	}
	// }
	// *confirmation = ""
	// *confirming = false

	file, err := os.Open("./files/stored/" + filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set content type
	contentType := "application/octet-stream"
	switch {
	case filename[len(filename)-4:] == ".txt":
		contentType = "text/plain"
	case filename[len(filename)-5:] == ".json":
		contentType = "application/json"
	case filename[len(filename)-4:] == ".mp4":
		contentType = "video/mp4"
	}

	// Set content disposition header
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", contentType)

	const chunkSize = 1024
	fmt.Println("File size: ")
	fmt.Println(stat.Size())
	if stat.Size() > chunkSize {
		fmt.Println("Must serve in chunks")
		buffer := make([]byte, chunkSize)
		for {
			//	time.Sleep(1 * time.Second)
			// Read 10 bytes from the file
			n, err := file.Read(buffer)
			// fmt.Println("n:", n)
			// fmt.Println("buffer:", buffer)
			// fmt.Println()
			if err != nil {
				// Check if it's the end of the file
				if err.Error() == "EOF" {
					break
				}
				// Handle other errors
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			fmt.Println("Sending chunk...")
			// Write the 10-byte chunk to the response
			w.Write(buffer[:n])
			<-eventChannel
			//w.Write([]byte("\n@@@@\n"))
		}
	} else {
		fmt.Println("sending in one piece")
		// Copy file contents to response body
		_, err = io.Copy(w, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}

	fmt.Printf("\nFile %s sent!\n> ", filename)
}

type FileData struct {
	FileName string `json:"filename"`
	Content  []byte `json:"content"`
}

func (server *HTTPServer) storeFile(w http.ResponseWriter, r *http.Request, confirming *bool, confirmation *string) {
	// Parse JSON object from request body
	var fileData FileData
	err := json.NewDecoder(r.Body).Decode(&fileData)
	if err != nil {
		http.Error(w, "Failed to parse JSON data", http.StatusBadRequest)
		return
	}

	// Ask for confirmation
	*confirming = true
	fmt.Printf("\nYou have just received a request to store file '%s'. Do you want to store the file? (yes/no): ", fileData.FileName)

	// Check if confirmation is received
	for *confirmation != "yes" {
		if *confirmation != "" {
			http.Error(w, fmt.Sprintf("Client declined to store file '%s'.", fileData.FileName), http.StatusUnauthorized)
			*confirmation = ""
			*confirming = false
			return
		}
	}
	*confirmation = ""
	*confirming = false

	// Create file
	file_hash, err := server.storage.PutFile(fileData.Content)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s", file_hash)
	fmt.Printf("\nStored file %s hash %s!\n> ", fileData.FileName, file_hash)
}

func getFile(w http.ResponseWriter, r *http.Request) {
	// Get the context from the request
	ctx := r.Context()

	// Check if the "filename" query parameter is present
	hasFilename := r.URL.Query().Has("filename")

	// Retrieve the value of the "filename" query parameter
	filename := r.URL.Query().Get("filename")

	// Print information about the request
	fmt.Printf("%s: got /file request. filename(%t)=%s\n",
		ctx.Value(keyServerAddr),
		hasFilename, filename,
	)

	// Check if the "filename" parameter is present
	if hasFilename {
		// Check if the file exists in the local directory
		filePath := filepath.Join(".", filename)
		if _, err := os.Stat(filePath); err == nil {
			// Serve the file using http.ServeFile
			http.ServeFile(w, r, filePath)
			fmt.Printf("Served %s to client\n", filename)
			return
		} else if os.IsNotExist(err) {
			// File not found
			http.Error(w, "File not found", http.StatusNotFound)
			return
		} else {
			// Other error
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		// Write a response indicating that no filename was found
		io.WriteString(w, "No filename found\n")
	}
}
func ConvertKeyToString(n *big.Int, e int) string {
	N := n // Replace with your actual modulus
	E := e // Replace with your actual public exponent

	// Create an RSA public key using the modulus and exponent
	publicKey := rsa.PublicKey{
		N: N,
		E: E,
	}

	// Marshal the public key into DER format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		panic(err)
	}

	// Create a PEM block for the public key
	publicKeyPEM := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	// Write the PEM-encoded public key to a file (or any io.Writer)
	publicKeyString := string(pem.EncodeToMemory(&publicKeyPEM))
	return publicKeyString
}
func jobRoutine(jobId string, hash string, peerId string) {
	holders, err := SetupCheckHolders(hash)
	if err != nil {
		fmt.Printf("Error finding holders for file: %x", err)
		return
	}
	var bestHolder *fileshare.User
	var selectedHolder *fileshare.User
	bestHolder = nil
	selectedHolder = nil
	for _, holder := range holders.Holders {
		if bestHolder == nil {
			bestHolder = holder
		} else if holder.GetPrice() < bestHolder.GetPrice() {
			bestHolder = holder
		}
		if string(holder.Id) == peerId {
			selectedHolder = holder
		}
	}
	if bestHolder == nil && selectedHolder == nil {
		fmt.Println("Unable to find holder for this hash.")
		return
	}
	if selectedHolder != nil {
		bestHolder = selectedHolder
	}
	fmt.Printf("%s - %d OrcaCoin\n", bestHolder.GetIp(), bestHolder.GetPrice())

	pubKeyInterface, err := x509.ParsePKIXPublicKey(bestHolder.Id)
	if err != nil {
		log.Fatal("failed to parse DER encoded public key: ", err)
	}
	rsaPubKey, ok := pubKeyInterface.(*rsa.PublicKey)
	if !ok {
		log.Fatal("not an RSA public key")
	}
	key := ConvertKeyToString(rsaPubKey.N, rsaPubKey.E)
	err = Client.GetFileOnce(bestHolder.GetIp(), bestHolder.GetPort(), hash, key, fmt.Sprintf("%d", bestHolder.GetPrice()), PassKey, jobId)
	if err != nil {
		fmt.Printf("Error getting file %s", err)
	}
}

type AddJobReqPayload struct {
	FileHash string `json:"fileHash"`
	PeerId   string `json:"peer"`
}

type AddJobResPayload struct {
	JobId string `json:"jobID"`
}

func AddJobHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		var payload AddJobReqPayload
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
			return
		}
		id := uuid.New()
		timeString := time.Now().Format(time.RFC3339)
		newJob := orcaJobs.Job{
			FileHash:        payload.FileHash,
			JobId:           id.String(),
			TimeQueued:      timeString,
			Status:          "active",
			AccumulatedCost: 0,
			ProjectedCost:   -1,
			ETA:             -1,
			PeerId:          payload.PeerId,
		}
		orcaJobs.AddJob(newJob)
		go jobRoutine(newJob.JobId, payload.FileHash, payload.PeerId)
		w.WriteHeader(http.StatusOK)
		response := AddJobResPayload{JobId: newJob.JobId}
		jsonData, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeStatusUpdate(w, "Failed to convert JSON Data into a string")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only PUT requests will be handled.")
		return
	}
}
