package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type FileData struct {
	FileName string `json:"filename"`
	Content  []byte `json:"content"`
}

func ImportFile(filePath string) {
	// Extract filename from the provided file path
	_, fileName := filepath.Split(filePath)
	if fileName == "" {
		fmt.Print("\nProvided path is a directory, not a file\n\n> ")
		return
	}

	// Open the source file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Print("\nFile does not exist\n\n> ")
		return
	}
	defer file.Close()

	// Create the directory if it doesn't exist
	err = os.MkdirAll("./files", 0755)
	if err != nil {
		return
	}

	// Save the file to the destination directory with the same filename
	destinationPath := filepath.Join("./files", fileName)
	destinationFile, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer destinationFile.Close()

	// Copy the contents of the source file to the destination file
	_, err = io.Copy(destinationFile, file)
	if err != nil {
		return
	}

	fmt.Printf("\nFile '%s' imported successfully!\n\n> ", fileName)
}

func GetFileOnce(ip, port, filename string) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%s/requestFile/%s", ip, port, filename))
	if err != nil {
		fmt.Printf("Error: %s\n\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}
		fmt.Printf("\nError: %s\n> ", body)
		return
	}

	// Create the directory if it doesn't exist
	err = os.MkdirAll("./files/requested/", 0755)
	if err != nil {
		return
	}

	// Create file
	out, err := os.Create("./files/requested/" + filename)
	if err != nil {
		return
	}
	defer out.Close()

	// Write response body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return
	}

	fmt.Printf("\nFile %s downloaded successfully!\n\n> ", filename)
}

func RequestStorage(ip, port, filename string) {
	// Read file content
	content, err := os.ReadFile("./files/documents/" + filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Create FileData struct
	fileData := FileData{
		FileName: filename,
		Content:  content,
	}

	// Marshal FileData to JSON
	jsonData, err := json.Marshal(fileData)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	// Send POST request to store file
	resp, err := http.Post(fmt.Sprintf("http://%s:%s/storeFile/", ip, port), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}
		fmt.Printf("\nError: %s\n> ", body)
		return
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	fmt.Println(string(body))
	fmt.Print("> ")
}
