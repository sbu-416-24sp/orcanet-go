package tests

import (
	orcaSearch "orca-peer/internal/search"
	"testing"
)

func TestSearchTxtFile(t *testing.T) {
	res, err := orcaSearch.SearchForFile("../../files", "tester.txt")
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if res != true {
		t.Errorf("Expected to find file, returned no file found")
	}
}

func TestSearchMP4File(t *testing.T) {
	res, err := orcaSearch.SearchForFile("../../files", "test.mp4")
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if res != true {
		t.Errorf("Expected to find file, returned no file found")
	}
}

func TestSearchFileNotFound(t *testing.T) {
	res, err := orcaSearch.SearchForFile("../../files", "tester.mp4")
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if res == true {
		t.Errorf("Expected to not find file, returned file found")
	}
}

func TestSearchErrorInvalidDirectory(t *testing.T) {
	_, err := orcaSearch.SearchForFile("filess", "tester.mp4")
	if err == nil {
		t.Errorf("Expected error: no directoyr found")
	}
}
func TestSearchErrorGetAllFilesBasic(t *testing.T) {
	_, err := orcaSearch.GetAllFiles("files")
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
}
