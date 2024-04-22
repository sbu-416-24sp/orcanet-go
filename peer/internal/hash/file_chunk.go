package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"orca-peer/internal/fileshare"
	"os"
	"strings"
)

type FileChunk struct {
	Hashes    []string
	BytesRead int64
}

func GetFileKey(filePath string, fileName string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	chunkSize := 4 * 1024 * 1024
	chunk := make([]byte, chunkSize)

	hasher := sha256.New()
	hashedFiles := FileChunk{}
	for {
		bytesRead, err := file.Read(chunk)
		if err != nil && err != io.EOF {
			return "", err
		}
		if bytesRead == 0 {
			break
		}
		hasher.Write(chunk[:bytesRead])
		hash := hasher.Sum(nil)
		hashedFiles.Hashes = append(hashedFiles.Hashes, hex.EncodeToString(hash))
		hashedFiles.BytesRead += int64(bytesRead)
		hasher.Reset()
	}
	fileKey := fileshare.FileInfo{}
	fileKey.ChunkHashes = hashedFiles.Hashes
	fileKey.FileSize = hashedFiles.BytesRead
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	fileKey.FileHash = string(hasher.Sum(nil))
	fileKey.FileName = fileName
	concatKey := fileKey.FileHash + strings.Join(fileKey.ChunkHashes, "") + fmt.Sprint(fileKey.FileSize) + fileKey.FileName
	hasher.Reset()
	hasher.Write([]byte(concatKey))
	finalHashedKey := hex.EncodeToString(hasher.Sum(nil))
	return finalHashedKey, nil
}
