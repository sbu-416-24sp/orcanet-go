package api

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"orca-peer/internal/fileshare"
	orcaHash "orca-peer/internal/hash"
	orcaJobs "orca-peer/internal/jobs"
	orcaMining "orca-peer/internal/mining"
	"orca-peer/internal/server"
	"os"
	"strconv"
	"strings"
)

type UploadFileJSONBody struct {
	Filepath string `json:"filepath"`
	fileData http.File
}

var backend *Backend
var peers *PeerStorage
var publicKey *rsa.PublicKey
var privateKey *rsa.PrivateKey
var storedFileInfoMap map[string]fileshare.FileInfo

type GetFileJSONResponseBody struct {
	Filename    string   `json:"name"`
	Size        int      `json:"size"`
	NumberPeers int      `json:"numberOfPeers"`
	Producers   []string `json:"listProducers"`
}

func getFile(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	queryParams := r.URL.Query()
	hash := queryParams.Get("hash")
	chunkIndex := queryParams.Get("chunk-index")

	// Check if the "hash" parameter is present
	if hash == "" {
		http.Error(w, "Missing 'hash' parameter", http.StatusBadRequest)
		return
	}
	fmt.Println("hash:", hash)
	fileaddress := ""

	if chunkIndex == "" {
		http.Error(w, "Missing 'chunk-index' parameter", http.StatusBadRequest)
		return
	}
	fmt.Println("chunk:", chunkIndex)
	chunkIndexInt, err := strconv.Atoi(chunkIndex)
	if err != nil {
		http.Error(w, "Bad chunk index parameter", http.StatusBadRequest)
		return
	}
	fileaddress = ""
	orcaFileInfo, ok := storedFileInfoMap[hash]
	if !ok {
		http.Error(w, "Specified hash is not in orcastore fileshare server node list", http.StatusBadRequest)
	}

	hashes := orcaFileInfo.ChunkHashes
	if chunkIndexInt >= len(hashes) {
		http.Error(w, "Bad chunk index parameter", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat("files/stored/" + hashes[chunkIndexInt]); !os.IsNotExist(err) {
		fileaddress = "files/stored/" + hashes[chunkIndexInt]
	}

	if fileaddress != "" {
		fmt.Println("File address:", fileaddress)
		w.Header().Set("X-Chunks-Length", fmt.Sprintf("%d", len(hashes)))
		http.ServeFile(w, r, fileaddress)

	} else if os.IsNotExist(err) {
		w.WriteHeader(http.StatusBadRequest)
		writeStatusUpdate(w, "File hash does not exist in directory.")
		return
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		writeStatusUpdate(w, "Error arose checking for file.")
		return
	}
	w.WriteHeader(http.StatusOK)
	writeStatusUpdate(w, "Successfully removed file.")
	return
}
func getAllFiles(w http.ResponseWriter, r *http.Request) {

}

type FileInfo struct {
	Filename     string `json:"filename"`
	Filesize     int    `json:"filesize"`
	Filehash     string `json:"filehash"`
	Lastmodified string `json:"lastmodified"`
}

func writeStatusUpdate(w http.ResponseWriter, message string) {
	responseMsg := map[string]interface{}{
		"status": message,
	}
	responseMsgJsonString, err := json.Marshal(responseMsg)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseMsgJsonString)
}

type HashResponse struct {
	Hash string `json:"hash"`
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		sourceFile, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Unable to get file from form", http.StatusBadRequest)
			return
		}
		fileContent, err := io.ReadAll(sourceFile)
		if err != nil {
			http.Error(w, "Unable to read file", http.StatusInternalServerError)
			return
		}
		defer sourceFile.Close()
		hash := sha256.Sum256(fileContent)
		hexHash := hex.EncodeToString(hash[:])
		destinationFilePath := "./files/" + hexHash
		destinationFile, err := os.Create(destinationFilePath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeStatusUpdate(w, "Cannot create the file to store base64 data.")
			return
		}
		defer destinationFile.Close()
		_, err = sourceFile.Seek(0, 0)
		if err != nil {
			http.Error(w, "Unable to reset file read cursor", http.StatusInternalServerError)
			return
		}
		_, err = io.Copy(destinationFile, sourceFile)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeStatusUpdate(w, "Unable to copy base64 data.")
			return
		}
		fmt.Fprintf(w, "File %s uploaded successfully\n", handler.Filename)
		// Create a JSON response containing the hash
		response := HashResponse{Hash: hexHash}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Unable to marshal JSON", http.StatusInternalServerError)
			return
		}
		writeStatusUpdate(w, "Successfully uploaded file from local computer into files directory")
		w.Write(jsonResponse)
		return
	}
}

type WriteFileJSONBody struct {
	Base64File       string `json:"base64File"`
	Filesize         string `json:"fileSize"`
	OriginalFileName string `json:"originalFileName"`
}

func writeFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			var payload WriteFileJSONBody
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
				return
			}
			backend.UploadFile(payload.Base64File, payload.OriginalFileName, payload.Filesize)
		default:
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Request must have the content header set as application/json")
			return
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only POST requests will be handled.")
		return
	}

}

func handleFileRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		path := r.URL.Path
		parts := strings.Split(path, "/")
		if len(parts) != 3 {
			http.NotFound(w, r)
			return
		}
		hash := parts[2]
		filePath := "./files/" + hash
		if _, err := os.Stat(filePath); err == nil {
			err := os.Remove(filePath)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				writeStatusUpdate(w, "Error removing file from local directory.")
				return
			}

		} else if os.IsNotExist(err) {
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "File hash does not exist in directory.")
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			writeStatusUpdate(w, "Error arose checking for file.")
			return
		}
		w.WriteHeader(http.StatusOK)
		writeStatusUpdate(w, "Successfully removed file.")
		return
	} else if r.Method == http.MethodGet {
		path := r.URL.Path
		parts := strings.Split(path, "/")
		if len(parts) != 4 || parts[3] != "info" {
			http.NotFound(w, r)
			return
		}
		hash := parts[2]
		holders, err := server.SetupCheckHolders(hash)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Unable to find holders of this file.")
		}
		peers := make([]string, 0)
		for _, holder := range holders.Holders {
			peers = append(peers, holder.Ip)
		}
		responseBody := GetFileJSONResponseBody{
			Filename:    hash,
			Size:        0,
			NumberPeers: len(peers),
			Producers:   peers,
		}
		jsonData, err := json.Marshal(responseBody)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeStatusUpdate(w, "Failed to convert JSON Data into a string")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only DELETE requests will be handled.")
		return
	}
}

func InitServer(fileInfoMap *map[string]fileshare.FileInfo) {
	storedFileInfoMap = *fileInfoMap
	backend = NewBackend()
	peers = NewPeerStorage()
	fmt.Println("Settig up API Routes")
	publicKey, privateKey = orcaHash.LoadInKeys()
	orcaJobs.InitJobRoutes()
	orcaMining.InitDeviceTracker()
	http.HandleFunc("/file/", handleFileRoute)
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/get-file", getFile)

	http.HandleFunc("/writeFile", writeFile)
	http.HandleFunc("/sendMoney", sendMoney)
	http.HandleFunc("/getLocation", getLocation)
	http.HandleFunc("/job-peer", JobPeerHandler)
	http.HandleFunc("/device", orcaMining.PutDeviceHandler)
	http.HandleFunc("/device_list", orcaMining.PutDeviceHandler)
}
