package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"orca-peer/internal/fileshare"
	"os"
	"strings"
	"io/ioutil"
	"errors"
)

type FileChunk struct {
	Hashes    []string
	BytesRead int64
}

//Returns hash key, fileinfo struct, and error if any
//will write individual chunks to /files/stored
func SaveChunkedFile(filePath string, fileName string) (string, fileshare.FileInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fileshare.FileInfo{}, err
	}
	defer file.Close()
	chunkSize := 4 * 1024 * 1024
	chunk := make([]byte, chunkSize)

	hasher := sha256.New()
	hashedFiles := FileChunk{}
	for {
		bytesRead, err := file.Read(chunk)
		if err != nil && err != io.EOF {
			return "", fileshare.FileInfo{}, err
		}
		if bytesRead == 0 {
			break
		}

		hasher.Write(chunk[:bytesRead])
		hash := hasher.Sum(nil)
		hashedFiles.Hashes = append(hashedFiles.Hashes, hex.EncodeToString(hash))
		err = ioutil.WriteFile("./files/stored/" + hex.EncodeToString(hash), chunk, 0777)
		if err != nil {
			//clean up any written hashes
			for _, chunkHash := range hashedFiles.Hashes {
				err = os.Remove("./files/stored/" + chunkHash)
				if err != nil {
					return "", fileshare.FileInfo{}, errors.New(fmt.Sprintf("Failed to clean up removing partial chunks for error: %s", err))
				}
			}
			return "", fileshare.FileInfo{}, err
		}
		hashedFiles.BytesRead += int64(bytesRead)
		hasher.Reset()
	}
	fileKey := fileshare.FileInfo{}
	fileKey.ChunkHashes = hashedFiles.Hashes
	fileKey.FileSize = hashedFiles.BytesRead
	// if _, err := io.Copy(hasher, file); err != nil {
	// 	return "", nil, err
	// }
	fileKey.FileHash = string(hasher.Sum(nil))
	fileKey.FileName = fileName
	concatKey := fileKey.FileHash + strings.Join(fileKey.ChunkHashes, "") + fmt.Sprint(fileKey.FileSize) + fileKey.FileName
	hasher.Reset()
	hasher.Write([]byte(concatKey))
	finalHashedKey := hex.EncodeToString(hasher.Sum(nil))
	return finalHashedKey, fileKey, nil
}
