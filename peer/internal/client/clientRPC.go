package client

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	pb "orca-peer/internal/fileshare"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

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
