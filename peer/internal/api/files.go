package api

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"os"
	"sync" // Import sync for mutex
)

type Backend struct {
	ctx        context.Context
	mutex      sync.RWMutex // Protects the activities slice and counter
	activities []Activity   // The list of activities
	counter    int          // Global counter for IDs
}

func NewBackend() *Backend {
	return &Backend{
		ctx:        context.Background(), // Initialize context or accept it as a parameter if needed
		mutex:      sync.RWMutex{},       // Initialized mutex
		activities: make([]Activity, 0),  // Empty slice of activities
		counter:    0,                    // Starting counter value
	}
}
func generateHash(file io.Reader) string {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func (b *Backend) UploadFile(base64File string, originalFileName string, fileSize string) error {
	decodedFile, err := base64.StdEncoding.DecodeString(base64File)
	if err != nil {
		return err
	}

	tempFile, err := os.CreateTemp("", "files/"+originalFileName)
	if err != nil {
		return err
	}
	defer tempFile.Close()

	_, err = tempFile.Write(decodedFile)
	if err != nil {
		return err
	}

	// Reset file pointer for hash generation
	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	b.mutex.Lock() // Ensure exclusive access to modify activities and counter
	defer b.mutex.Unlock()

	hash := generateHash(tempFile)

	created_activity := Activity{ // Directly creating an Activity object
		Name:   originalFileName,
		Size:   fileSize,
		Hash:   hash,
		Status: "Uploaded",
		Peers:  0,
		ID:     b.counter, // Use the counter within Backend
	}

	b.activities = append(b.activities, created_activity)
	b.counter++ // Safely increment the counter within the lock

	return nil
}
