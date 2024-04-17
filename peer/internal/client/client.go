package client

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"orca-peer/internal/hash"
	orcaHash "orca-peer/internal/hash"
	"os"
	"path/filepath"
)

type Client struct {
	name_map hash.NameMap
}

func NewClient(path string) *Client {
	return &Client{
		name_map: *hash.NewNameStore(path),
	}
}

type FileData struct {
	FileName string `json:"filename"`
	Content  []byte `json:"content"`
}

func (client *Client) ImportFile(filePath string) error {
	// Extract filename from the provided file path
	_, fileName := filepath.Split(filePath)
	if fileName == "" {
		return fmt.Errorf("Provided path is a directory, not a file")
	}

	// Open the source file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Print("\nFile does not exist\n> ")
		return err
	}
	defer file.Close()

	// Create the directory if it doesn't exist
	err = os.MkdirAll("./files", 0755)
	if err != nil {
		return err
	}

	// Save the file to the destination directory with the same filename
	destinationPath := filepath.Join("./files", fileName)
	destinationFile, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// Copy the contents of the source file to the destination file
	_, err = io.Copy(destinationFile, file)
	if err != nil {
		return err
	}

	fmt.Printf("\nFile '%s' imported successfully!\n> ", fileName)
	return nil
}

type Data struct {
	Bytes               []byte `json:"bytes"`
	UnlockedTransaction []byte `json:"transaction"`
	PublicKey           string `json:"public_key"`
}

func SendTransaction(price float64, ip string, port string, publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey) {
	cost := orcaHash.GeneratePriceBytes(price)
	byteBuffer := bytes.NewBuffer(cost)
	pubKeyString, err := orcaHash.ExportRsaPublicKeyAsPemStr(publicKey)
	if err != nil {
		fmt.Println("Error sending public key in header:", err)
		return
	}
	data := Data{
		Bytes:               byteBuffer.Bytes(),
		UnlockedTransaction: cost,
		PublicKey:           string(pubKeyString),
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		os.Exit(1)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%s/sendTransaction", ip, port), bytes.NewReader(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	fmt.Println("Verifying Signature...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	} else {
		fmt.Println("Send Request")
	}
	defer resp.Body.Close()

}
func (client *Client) GetFileOnce(ip, port, filename string) error {
	/*
		file_hash := client.name_map.GetFileHash(filename)
		if file_hash == "" {
			fmt.Println("Error: do not have hash for the file")
			return
		}
	*/
	data, err := client.getData(ip, port, filename)
	if err != nil {
		return err
	}
	resp, err := http.Get(fmt.Sprintf("http://%s:%s/requestFile/%s", ip, port, filename))
	if err != nil {
		return err
	}

	fmt.Println("Response:")
	fmt.Println(resp)
	fmt.Println("ResponseBody:")
	fmt.Println(resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return err
		}
		fmt.Printf("\nError: %s\n> ", body)
		return err
	}

	// Create the directory if it doesn't exist
	err = os.MkdirAll("./files/requested/", 0755)
	if err != nil {
		panic(err)
	}

	// Create file
	_, err = os.Create("./files/requested/" + filename)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join("./files/requested/", filename), data, 0666)
	if err != nil {
		return err
	}

	fmt.Printf("\nFile %s downloaded successfully!\n> ", filename)
	return nil
}

func (client *Client) RequestStorage(ip, port, filename string) (string, error) {
	// Read file content
	content, err := os.ReadFile("./files/requested/" + filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return "", err
	}

	// Create FileData struct
	fileData := FileData{
		FileName: filename,
		Content:  content,
	}
	hash, err := client.storeData(ip, port, filename, &fileData)

	fmt.Print("> ")

	return hash, err
}

func (client *Client) GetDirectory(ip, port, path string) {
	data, err := client.getData(ip, port, path)
	if err != nil {
		fmt.Println("Failed to Get Directory")
		return
	}
	var dir_tree map[string]any
	err = json.Unmarshal(data, &dir_tree)
	if err != nil {
		fmt.Println("Failed to parse dir tree")
		return
	}
	err = client.getDirectory(ip, port, dir_tree)
	if err != nil {
		fmt.Println("Failed to Get Directory")
		return
	}
}

func (client *Client) getDirectory(ip, port string, dir_tree map[string]any) error {
	for path, v := range dir_tree {
		switch val := v.(type) {
		case string:
			err := os.MkdirAll(filepath.Join("./files/requested/", filepath.Dir(path)), 0755)
			if err != nil {
				return err
			}
			err = client.GetFileOnce(ip, port, path)
			if err != nil {
				return err
			}
		case map[string]any:
			client.getDirectory(ip, port, val)
		default:
			panic("Bug: dir_tree should only have strings or recursive dir_tree")
		}
	}
	return nil
}

func (client *Client) StoreDirectory(ip, port, path string) {
	dir_tree_hashes, err := client.storeDirectory(ip, port, filepath.Join("./files/documents/", path))
	if err != nil {
		fmt.Println("Error storing directory", path)
	}
	data, err := json.Marshal(dir_tree_hashes)
	if err != nil {
		fmt.Println("Error parsing directory hash tree")
	}
	filedata := FileData{
		FileName: path,
		Content:  data,
	}
	dir_hash, err := client.storeData(ip, port, path, &filedata)
	if err != nil {
		fmt.Println("Error storing directory", path)
		return
	}
	client.name_map.PutFileHash(path, dir_hash)
}

func (client *Client) storeDirectory(ip, port string, path string) (map[string]any, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Println("Error reading directory", path)
		return nil, err
	}
	mapping := map[string]any{}

	for _, entry := range entries {
		path := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			sub_mapping, err := client.storeDirectory(ip, port, path)
			if err != nil {
				return nil, err
			}
			mapping[path] = sub_mapping
		} else {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			filedata := FileData{
				FileName: path,
				Content:  data,
			}

			file_hash, err := client.storeData(ip, port, path, &filedata)
			if err != nil {
				return nil, err
			}
			mapping[path] = file_hash
		}
	}
	return mapping, nil
}

func (client *Client) storeData(ip, port, filename string, fileData *FileData) (string, error) {
	// Marshal FileData to JSON
	jsonData, err := json.Marshal(fileData)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return "", err
	}

	// Send POST request to store file
	resp, err := http.Post(fmt.Sprintf("http://%s:%s/storeFile/", ip, port), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return "", err
		}
		fmt.Printf("\nError: %s\n> ", body)
		return "", errors.New("http status not ok")
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "", err
	}
	client.name_map.PutFileHash(filename, string(body))

	fmt.Println(string(body))
	return string(body), nil
}

func (client *Client) getData(ip, port, filename string) ([]byte, error) {

	file_hash := client.name_map.GetFileHash(filename)
	if file_hash == "" {
		fmt.Println("Error: do not have hash for the file")
		return nil, errors.New("name not found")
	}
	resp, err := http.Get(fmt.Sprintf("http://%s:%s/requestFile/%s", ip, port, file_hash))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return nil, err
		}
		fmt.Printf("\nError: %s\n> ", body)
		return nil, errors.New("http status not ok")
	}

	data := bytes.NewBuffer([]byte{})

	_, err = io.Copy(data, resp.Body)
	if err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}
