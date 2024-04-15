package test

import (
	"fmt"
	"log"
	orcaClient "orca-peer/internal/client"
	pb "orca-peer/internal/fileshare"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type testFilePeerServer struct {
	pb.UnimplementedFileShareServer
	savedAddress map[string][]*pb.StorageIP
	savedFiles   []*pb.FileDesc // read-only after initialized

	mu sync.Mutex // protects routeNotes
}

func RunTestServer() {
	serverIP := "localhost:50051"
	go SetupTestMarket()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial(serverIP, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewFileShareClient(conn)
	orcaClient.RequestFileFromMarket(client, &pb.CheckHoldersRequest{})

	blockchainIP := "localhost:50052"
	go SetupTestBlockchain()
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err = grpc.Dial(blockchainIP, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	orcaClient.RecordTransactionWrapper(client, &pb.FileRequestTransaction{FileByteSize: 100,
		FileHashName:      "abc",
		CurrencyExchanged: float32(1),
		SenderId:          "s1",
		ReceiverId:        "r1",
		FileIpLocation:    "localhost:50051",
		SecondsTimeout:    100,
	})
	for {
		fmt.Println("Waiting...")
		time.Sleep(10 * time.Second) // or runtime.Gosched() or similar per @misterbee
	}
}
