package store

import (
	"context"
	"fmt"
	"io"
	"log"
	pb "orca-peer/internal/fileshare"
	"os"
	"time"
)

type FileInfo struct {
	IsDir   bool
	ModTime time.Time
	Name    string
	Size    int64
}

func GetAllLocalFiles() []FileInfo {
	files, err := os.ReadDir("files/stored")
	if err != nil {
		log.Fatal(err)
	}
	fileNames := make([]FileInfo, 1)
	for _, file := range files {
		fileInfo, err := os.Stat("files/stored/" + file.Name())
		if err == nil {
			fileNames = append(fileNames, FileInfo{IsDir: fileInfo.IsDir(), ModTime: fileInfo.ModTime(), Name: fileInfo.Name(), Size: fileInfo.Size()})
		}
	}
	return fileNames
}

func GetAllMarketFiles(client pb.FileShareClient, me *pb.StorageIP) []*pb.FileDesc {
	log.Printf("Requesting All File Names")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	streamOfFiles, err := client.RequestAllAvailableFileNames(ctx, me)
	if err != nil {
		fmt.Println("Error requesting files from stream")
	}
	var all_files = []*pb.FileDesc{}
	for {
		file_desc, err := streamOfFiles.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("client.RequestAllAvailableFileNames failed: %v", err)
		}
		log.Printf("File named %s found with size %d for price %f ",
			file_desc.FileNameHash, file_desc.FileBytes, file_desc.FileCost)
		all_files = append(all_files, file_desc)
	}
	return all_files
}
