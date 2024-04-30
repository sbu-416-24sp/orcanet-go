package client

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"context"
	"log"
	"net/http"
	orcaBlockchain "orca-peer/internal/blockchain"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"orca-peer/internal/hash"
	orcaHash "orca-peer/internal/hash"
	orcaJobs "orca-peer/internal/jobs"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
	"os"
	"encoding/binary"
	"bufio"
	"path/filepath"
	"strconv"
	"time"
)

type Client struct {
	name_map   hash.NameMap
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
	Host host.Host
}

func NewClient(path string) *Client {
	return &Client{
		name_map:   *hash.NewNameStore(path),
		PublicKey:  nil,
		PrivateKey: nil,
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
		return errors.New("directory given, not file")
	}

	src, err := os.Open(filePath)
	if err != nil {
		return errors.New("cant find given absolute file path")
	}
	defer src.Close()
	destinationFile, err := os.Create("./files/" + fileName)
	if err != nil {
		return errors.New("error creating destination file")
	}
	defer destinationFile.Close()
	_, err = io.Copy(destinationFile, src)
	if err != nil {
		return errors.New("error copying file")
	}
	fmt.Println("Sucessfully imported file")
	return nil
}

type Data struct {
	Bytes               []byte  `json:"bytes"`
	UnlockedTransaction []byte  `json:"transaction"`
	PublicKey           string  `json:"public_key"`
	Date                string  `json:"date"`
	Cost                float64 `json:"cost"`
}

func SendTransaction(price float64, ip string, port string, publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey) {
	cost := orcaHash.GeneratePriceBytes(price)
	byteBuffer := bytes.NewBuffer(cost)
	pubKeyString, err := orcaHash.ExportRsaPublicKeyAsPemStr(publicKey)
	if err != nil {
		fmt.Println("Error sending public key in header:", err)
		return
	}
	currentTime := time.Now()
	dateTimeString := currentTime.Format(time.RFC3339Nano)
	data := Data{
		Bytes:               byteBuffer.Bytes(),
		UnlockedTransaction: cost,
		PublicKey:           string(pubKeyString),
		Date:                dateTimeString,
		Cost:                price,
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
	err = os.WriteFile("./files/transactions/"+dateTimeString, jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing transaction to file:", err)
		return
	}
}
func (client *Client) GetFileOnce(ip string, port int32, file_hash string, walletAddress string, price string, passKey string, jobId string) error {
	//Dial peer and start stream to request file
	peerMA, err := multiaddr.NewMultiaddr(ip)
	if err != nil {
		log.Println(err)
		orcaJobs.UpdateJobStatus(jobId, "terminated")
		return err
	}

	peer, err := peer.AddrInfoFromP2pAddr(peerMA)
	if err != nil {
		log.Println(err)
		orcaJobs.UpdateJobStatus(jobId, "terminated")
		return err
	}

	client.Host.Peerstore().AddAddrs(peer.ID, peer.Addrs, peerstore.AddressTTL)

	err = client.Host.Connect(context.Background(), *peer)
	if err != nil {
		log.Println(err)
		orcaJobs.UpdateJobStatus(jobId, "terminated")
		return err
	}

	s, err := client.Host.NewStream(context.Background(), peer.ID, protocol.ID("orcanet-fileshare/1.0/" + file_hash))
	if err != nil {
		log.Println(err)
		orcaJobs.UpdateJobStatus(jobId, "terminated")
		return err
	}
	defer s.Close()

	//continously send request and process response from peer
	chunkIndex := -1
	for {
		fileChunkReq := orcaJobs.FileChunkRequest{
			FileHash: file_hash,
			ChunkIndex: chunkIndex + 1,
			JobId: jobId,
		}
	
		nextChunkReqBytes, err := json.Marshal(fileChunkReq)
		if err != nil {
			fmt.Println("Error:", err)
			orcaJobs.UpdateJobStatus(jobId, "terminated")
			return err
		}
	
		lengthBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(lengthBytes, uint32(len(nextChunkReqBytes)))
		_, err = s.Write(lengthBytes)
		if err != nil {
			fmt.Println(err)
			orcaJobs.UpdateJobStatus(jobId, "terminated")
			return nil
		}
		
		_, err = s.Write(nextChunkReqBytes)
		if err != nil {
			fmt.Println(err)
			orcaJobs.UpdateJobStatus(jobId, "terminated")
			return nil
		}

		buf := bufio.NewReader(s)
		lengthBytes = make([]byte, 0)
		for i := 0; i < 4; i++ {
			b, err := buf.ReadByte()
			if err != nil {
				fmt.Println(err)
				orcaJobs.UpdateJobStatus(jobId, "terminated")
				return err
			}	
			lengthBytes = append(lengthBytes, b)
		}

		length := binary.LittleEndian.Uint32(lengthBytes)
		payload := make([]byte, length)
		_, err = io.ReadFull(buf, payload)
		if err != nil {
			fmt.Println(err)
			orcaJobs.UpdateJobStatus(jobId, "terminated")
			return err
		}
		
		fileChunk := orcaJobs.FileChunk{}
		err = json.Unmarshal(payload, &fileChunk)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			orcaJobs.UpdateJobStatus(jobId, "terminated")
			return err
		}
		hash := fileChunk.FileHash

		err = client.sendTransactionFee(price, walletAddress, passKey)
		if err != nil {
			orcaJobs.UpdateJobStatus(jobId, "terminated")
			return err
		}
		priceInt, err := strconv.ParseInt(price, 10, 64)
		if err != nil {
			fmt.Println(err)
		} else {
			if client.PublicKey != nil && client.PrivateKey != nil {
				SendTransaction(float64(priceInt), ip, string(port), client.PublicKey, client.PrivateKey)
			}
			orcaJobs.UpdateJobCost(jobId, int(priceInt))
		}

		file, err := os.OpenFile("./files/requested/" + hash, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		defer file.Close()
	
		_, err = file.Write(fileChunk.Data)
		if err != nil {
			log.Fatal(err)
			orcaJobs.UpdateJobStatus(jobId, "terminated")
			return err
		}

		fmt.Printf("Chunk %d for %s received and written\n", hash, fileChunk.ChunkIndex)

		if fileChunk.ChunkIndex == fileChunk.MaxChunk - 1 {
			fmt.Println("All chunks received and written")
			break
		}

		chunkIndex += 1

		if jobId != "" {
			status := orcaJobs.GetJobStatus(jobId)
			if status == "terminated" {
				return nil
			} else if status == "paused" {
				for {
					time.Sleep(10 * time.Second)
					if orcaJobs.GetJobStatus(jobId) != "paused" {
						break
					}
				}
			}
		}
	}

	orcaJobs.UpdateJobStatus(jobId, "finished")
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

func (client *Client) GetDirectory(ip string, port int32, path string) {
	// data, err := client.getData(ip, port, path)
	// if err != nil {
	// 	fmt.Println("Failed to Get Directory")
	// 	return
	// }
	// var dir_tree map[string]any
	// err = json.Unmarshal(data, &dir_tree)
	// if err != nil {
	// 	fmt.Println("Failed to parse dir tree")
	// 	return
	// }
	// err = client.getDirectory(ip, port, dir_tree)
	// if err != nil {
	// 	fmt.Println("Failed to Get Directory")
	// 	return
	// }
}

func (client *Client) getDirectory(ip string, port int32, dir_tree map[string]any) error {
	for path, v := range dir_tree {
		switch val := v.(type) {
		case string:
			err := os.MkdirAll(filepath.Join("./files/requested/", filepath.Dir(path)), 0755)
			if err != nil {
				return err
			}
			// need to fix to match new blockchain requirements
			err = client.GetFileOnce(ip, port, path, "", "", "", "")
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

func (client *Client) sendTransactionFee(coins string, address string, senderWalletPass string) error {
	err := orcaBlockchain.SendToAddress(coins, address, senderWalletPass)
	return err
}

// int return value will be the length of chunk indexes from response header
func (client *Client) getChunkData(ip string, port int32, file_hash string, chunk int) (int, []byte, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/get-file?hash=%s&chunk-index=%d", ip, port, file_hash, chunk))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return -1, nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return -1, nil, err
		}
		fmt.Printf("\nError: %s\n ", body)
		return -1, nil, errors.New("http status not ok")
	}

	data := bytes.NewBuffer([]byte{})

	_, err = io.Copy(data, resp.Body)
	if err != nil {
		return -1, nil, err
	}

	chunkLengths, err := strconv.Atoi(resp.Header.Get("X-Chunks-Length"))
	if err != nil {
		return -1, nil, err
	}

	return chunkLengths, data.Bytes(), nil
}

// func (client *Client) getData(ip string, port int32, file_hash string) ([]byte, error) {

// 	// file_hash := client.name_map.GetFileHash(filename)
// 	// if file_hash == "" {
// 	// 	fmt.Println("Error: do not have hash for the file")
// 	// 	return nil, errors.New("name not found")
// 	// }
// 	resp, err := http.Get(fmt.Sprintf("http://%s:%d/get-file?hash=%s&chunk=0", ip, port, file_hash))
// 	if err != nil {
// 		fmt.Printf("Error: %s\n", err)
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		body, err := io.ReadAll(resp.Body)
// 		if err != nil {
// 			fmt.Println("Error reading response body:", err)
// 			return nil, err
// 		}
// 		fmt.Printf("\nError: %s\n ", body)
// 		return nil, errors.New("http status not ok")
// 	}

// 	data := bytes.NewBuffer([]byte{})

// 	_, err = io.Copy(data, resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return data.Bytes(), nil
// }
