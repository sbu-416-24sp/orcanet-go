package test

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "orca-peer/internal/fileshare"

	"google.golang.org/grpc"
)

func (s *testFilePeerServer) RecordFileRequestTransaction(ctx context.Context, transaction *pb.FileRequestTransaction) (*pb.TransactionACKResponse, error) {
	return &pb.TransactionACKResponse{IsSuccess: true, BlockHash: "abc", Timestamp: 10.0, MarketId: "Market1"}, nil
}
func newBlockchainServer() *testFilePeerServer {
	s := &testFilePeerServer{}
	return s
}

func SetupTestBlockchain() {
	port := 50052
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterFileShareServer(grpcServer, newBlockchainServer())
	grpcServer.Serve(lis)
}
