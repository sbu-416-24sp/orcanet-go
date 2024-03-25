package test

import (
	"fmt"
	"log"
	"net"

	pb "orca-peer/internal/fileshare"

	"google.golang.org/grpc"
)

func (s *testFilePeerServer) RequestAllAvailableFileNames(address *pb.StorageIP, stream pb.FileShare_RequestAllAvailableFileNamesServer) error {
	for _, feature := range s.savedFiles {
		if err := stream.Send(feature); err != nil {
			return err
		}
	}
	return nil
}
func newServer() *testFilePeerServer {
	s := &testFilePeerServer{}
	s.savedAddress = make(map[string][]*pb.StorageIP)
	s.savedFiles = make([]*pb.FileDesc, 0)
	s.savedAddress["file1"] = make([]*pb.StorageIP, 0)
	s.savedFiles = append(s.savedFiles, &pb.FileDesc{FileNameHash: "fABC", FileCost: 1000, FileSizeBytes: 75})
	s.savedFiles = append(s.savedFiles, &pb.FileDesc{FileNameHash: "fABCD", FileCost: 100, FileSizeBytes: 70})
	s.savedAddress["file1"] = append(s.savedAddress["file1"], &pb.StorageIP{Success: true, Address: "localhost:9876", UserId: "server-1", FileName: "Test1.txt", FileByteSize: 100, FileCost: 1.0, IsLastCandidate: false})
	s.savedAddress["file1"] = append(s.savedAddress["file1"], &pb.StorageIP{Success: true, Address: "localhost:9877", UserId: "server-1", FileName: "Test1.txt", FileByteSize: 100, FileCost: 2.0, IsLastCandidate: false})
	s.savedAddress["file1"] = append(s.savedAddress["file1"], &pb.StorageIP{Success: true, Address: "localhost:9876", UserId: "server-1", FileName: "Test1.txt", FileByteSize: 100, FileCost: 5.0, IsLastCandidate: true})
	return s
}

func SetupTestMarket() {
	port := 50051
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterFileShareServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
