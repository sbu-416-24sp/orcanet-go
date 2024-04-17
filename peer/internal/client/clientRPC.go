package client

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	pb "orca-peer/internal/fileshare"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func RequestFileFromMarket(client pb.FileShareClient, fileDesc *pb.CheckHoldersRequest) *pb.HoldersResponse {
	log.Printf("Requesting IP For File (%s)", fileDesc.FileHash)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	holders, err := client.CheckHolders(ctx, fileDesc)
	if err != nil {
		log.Fatalf("client.requestFileStorage failed: %v", err)
	}
	return holders
}

func RequestFileFromProducer(baseURL string, filename string) bool {
	encodedParams := url.Values{}
	encodedParams.Add("filename", filename)
	queryString := encodedParams.Encode()

	// Construct the URL with the query string
	urlWithQuery := fmt.Sprintf("%s?%s", baseURL, queryString)
	resp, err := http.Get(urlWithQuery)

	if err != nil {
		log.Fatalln(err)
	}
	//We Read the response body on the line below.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	//Convert the body to type string
	sb := string(body)
	fmt.Println(sb)
	return false
}
func DialRPC(serverAddr *string) pb.FileShareClient {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewFileShareClient(conn)
	return client
}

func runRecordTransaction(client pb.FileShareClient, transaction *pb.FileRequestTransaction) *pb.TransactionACKResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ackResponse, err := client.RecordFileRequestTransaction(ctx, transaction)
	if err != nil {
		log.Fatalf("client.RecordFileRequestTransaction failed: %v", err)
	}
	log.Printf("ACK Response: %v", ackResponse)
	return ackResponse
}

func RecordTransactionWrapper(client pb.FileShareClient, transaction *pb.FileRequestTransaction) {
	var ack = runRecordTransaction(client, transaction)
	if ack.IsSuccess {
		fmt.Printf("[Server]: Successfully recorded transaction in hash: %v", ack.BlockHash)
	} else {
		fmt.Println("[Server]: Unable to record transaction in blockchain")
	}
}
