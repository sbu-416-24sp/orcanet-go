package hash

import (
	"crypto/sha256"
	"io"
	"os"
)

type FileChunk struct {
	Hash      []byte
	BytesRead int
}

func GetChunkHash(filePath string) ([]FileChunk, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	chunkSize := 4 * 1024 * 1024
	chunk := make([]byte, chunkSize)

	hasher := sha256.New()
	var allChunks []FileChunk
	for {
		bytesRead, err := file.Read(chunk)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if bytesRead == 0 {
			break
		}
		hasher.Write(chunk[:bytesRead])
		hash := hasher.Sum(nil)

		currentChunk := FileChunk{
			Hash:      hash,
			BytesRead: bytesRead,
		}
		allChunks = append(allChunks, currentChunk)
		// Reset the hasher for the next chunk
		hasher.Reset()
	}
	return allChunks, nil
}
