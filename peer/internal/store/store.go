package store

import (
	"log"
	"os"
	"time"
)

type FileInfo struct {
	IsDir   bool
	ModTime time.Time
	Name    string
	Size    int64
	Hash    string
}

//Searches for files in stored. 
//Format of name in stored should be '64-char-sha256-hash [nickname]'
//File name will just be nick name instead of hash + nickname
func GetAllLocalFiles() []FileInfo {
	files, err := os.ReadDir("files/stored")
	if err != nil {
		log.Fatal(err)
	}
	fileNames := make([]FileInfo, 0)
	for _, file := range files {
		fileInfo, err := os.Stat("files/stored/" + file.Name())
		if err == nil{
			fileNames = append(fileNames, 
				FileInfo{IsDir: fileInfo.IsDir(), 
					ModTime: fileInfo.ModTime(), 
					Name: fileInfo.Name()[65:], 
					Size: fileInfo.Size(), 
					Hash: fileInfo.Name()[0:64]})
		}
	}
	return fileNames
}
