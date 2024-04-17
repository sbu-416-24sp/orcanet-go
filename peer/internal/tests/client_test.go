package tests

import (
	orcaClient "orca-peer/internal/client"
	"testing"
)

func TestImportMP4File(t *testing.T) {
	client := orcaClient.Client{}
	err := client.ImportFile("../../files/test.mp4")
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
}

func TestImportTxtFile(t *testing.T) {
	client := orcaClient.Client{}
	err := client.ImportFile("../../files/tester.txt")
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
}

func TestErrorDirectoryNotFile(t *testing.T) {
	client := orcaClient.Client{}
	err := client.ImportFile("../../files/")
	if err == nil {
		t.Errorf("Expected an error: Provided path is directory not file")
	}
}

func TestErrorFileNonExist(t *testing.T) {
	client := orcaClient.Client{}
	err := client.ImportFile("../../files/tester1.txt")
	if err == nil {
		t.Errorf("Expected an error: File does not exist")
	}
}
