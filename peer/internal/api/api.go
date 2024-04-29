package api

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	orcaCLI "orca-peer/internal/cli"
	"orca-peer/internal/fileshare"
	orcaHash "orca-peer/internal/hash"
	orcaJobs "orca-peer/internal/jobs"
	orcaMining "orca-peer/internal/mining"
	"orca-peer/internal/server"
	"os"
	"path/filepath"
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

type UploadFileReq struct {
	FilePath string `json:"filePath"`
	Price    int64  `json:"price"`
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var payload UploadFileReq
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
			return
		}
		fileName := filepath.Base(payload.FilePath)
		hashKey, _, err := orcaHash.SaveChunkedFile(payload.FilePath, fileName)
		if err != nil {
			http.Error(w, "Unable to create chunked file, maybe filepath doesnt exist?", http.StatusInternalServerError)
			return
		}

		sourceFile, err := os.Open(payload.FilePath)
		if err != nil {
			fmt.Println("Error opening source file:", err)
			return
		}
		defer sourceFile.Close()
		destinationFile, err := os.Create("./files/" + fileName)
		if err != nil {
			fmt.Println("Error creating destination file:", err)
			return
		}
		defer destinationFile.Close()
		_, err = io.Copy(destinationFile, sourceFile)
		if err != nil {
			fmt.Println("Error copying file:", err)
			return
		}

		err = server.SetupRegisterFile(payload.FilePath, fileName, payload.Price, orcaCLI.Ip, int32(orcaCLI.Port))
		if err != nil {
			http.Error(w, "Unable to store file on DHT", http.StatusInternalServerError)
			return
		}
		response := HashResponse{Hash: hashKey}
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

type GetFileJSONBody struct {
	Filename string `json:"filename"`
	Hash     string `json:"hash"`
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {

		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			// For JSON content type, decode the JSON into a struct
			var payload GetFileJSONBody
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
				return
			}
			if payload.Filename == "" && payload.Hash == "" {

				w.WriteHeader(http.StatusInternalServerError)
				writeStatusUpdate(w, "Missing Filename and CID values inside of the payload.")
				return
			}
			fileDir := "./files/"
			filePath := "./files/" + payload.Hash

			// Check if the file exists in the "stored" directory
			storedFilePath := filepath.Join(fileDir, "stored", payload.Hash)
			if _, err := os.Stat(storedFilePath); err == nil {
				//		filePath = storedFilePath
			}
			// Check if the file exists in the "requested" directory
			requestedFilePath := filepath.Join(fileDir, "requested", payload.Hash)
			if _, err := os.Stat(requestedFilePath); err == nil {
				filePath = requestedFilePath
			}
			fmt.Println("filePath: ", filePath)
			// Attempt to delete the file
			err := os.Remove(filePath)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				writeStatusUpdate(w, "Error removing file from local directory.")
				return
			}

			fmt.Println("File deleted successfully.")
			return

		default:
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Request must have the content header set as application/json")
			return
		}
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
	http.HandleFunc("/upload-file", uploadFile)
	http.HandleFunc("/delete-file", deleteFile)

	http.HandleFunc("/writeFile", writeFile)
	http.HandleFunc("/sendMoney", sendMoney)
	http.HandleFunc("/getLocation", getLocation)
	http.HandleFunc("/job-peer", JobPeerHandler)
	http.HandleFunc("/device", orcaMining.PutDeviceHandler)
	http.HandleFunc("/device_list", orcaMining.PutDeviceHandler)
}
