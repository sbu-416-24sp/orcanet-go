package tests

import (
	orcaHash "orca-peer/internal/hash"
	"testing"
)

func TestBasicHash(t *testing.T) {
	fileName := "test.mp4"
	_, err := orcaHash.HashFile(fileName)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
}

func TestErrorNoFile(t *testing.T) {
	fileName := "tester.mp4"
	_, err := orcaHash.HashFile(fileName)
	if err == nil {
		t.Errorf("Expected error: file not found")
	}
}
