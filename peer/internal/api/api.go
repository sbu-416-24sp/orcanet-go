package api

import (
	"crypto/rsa"
	"crypto/sha256"
	"strings"

	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	orcaHash "orca-peer/internal/hash"
	orcaJobs "orca-peer/internal/jobs"
	"orca-peer/internal/server"
	"os"
	"path/filepath"
)

type UploadFileJSONBody struct {
	Filepath string `json:"filepath"`
	fileData http.File
}

var backend *Backend
var peers *PeerStorage
var publicKey *rsa.PublicKey
var privateKey *rsa.PrivateKey

type GetFileJSONResponseBody struct {
	Filename    string   `json:"name"`
	Size        int      `json:"size"`
	NumberPeers int      `json:"numberOfPeers"`
	Producers   []string `json:"listProducers"`
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
func getAllFiles(w http.ResponseWriter, r *http.Request) {

}

type FileInfo struct {
	Filename     string `json:"filename"`
	Filesize     int    `json:"filesize"`
	Filehash     string `json:"filehash"`
	Lastmodified string `json:"lastmodified"`
}

func getFileInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		queryParams := r.URL.Query()

		// Retrieve specific query parameters by key
		filename := queryParams.Get("hash")

		if st, err := os.Stat("files/" + filename); !os.IsNotExist(err) {
			fileData, err := os.ReadFile("files/" + filename)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				writeStatusUpdate(w, "Failed to read in file from given path")
				return
			}
			lenData := len(fileData)

			hash := sha256.Sum256(fileData)

			// Encode the hash as a hexadecimal string
			hexHash := hex.EncodeToString(hash[:])

			fileInfoResp := FileInfo{
				Filename:     filename,
				Filesize:     lenData,
				Filehash:     hexHash,
				Lastmodified: st.ModTime().String(),
			}
			jsonData, err := json.Marshal(fileInfoResp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				writeStatusUpdate(w, "Failed to convert JSON Data into a string")
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			w.Write(jsonData)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Request must have the content header set as application/json")
		return
	}
}

func getAllStored(w http.ResponseWriter, r *http.Request) {

	var fileInfoList []FileInfo

	// Get a list of files in the directory
	dirPath := "./files/stored/"
	files, err := os.ReadDir(dirPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeStatusUpdate(w, "1 Failed to convert JSON Data into a string")
		return
	}

	// Iterate over each file
	for _, file := range files {
		// Construct the file path
		filePath := filepath.Join(dirPath, file.Name())

		fileData, err := os.ReadFile(filePath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeStatusUpdate(w, "Failed to read in file from given path")
			return
		}
		st, err := os.Stat(filePath)
		lenData := len(fileData)
		//	base64Encode := base64.StdEncoding.EncodeToString(fileData)
		hash := sha256.Sum256(fileData)

		// Encode the hash as a hexadecimal string
		hexHash := hex.EncodeToString(hash[:])

		fileInfoResp := FileInfo{
			Filename:     file.Name(),
			Filesize:     lenData,
			Filehash:     hexHash,
			Lastmodified: st.ModTime().String(),
			//		Filecontent:  base64Encode,
		}
		//jsonData, err := json.Marshal(fileInfoResp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeStatusUpdate(w, "Failed to convert JSON Data into a string")
			return
		}
		// Append FileInfo to the list
		fileInfoList = append(fileInfoList, fileInfoResp)
	}
	jsonData, err := json.Marshal(fileInfoList)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeStatusUpdate(w, "Failed to convert JSON Data into a string")
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
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
		// fileData := payload.fileData
		// sourceFile, err := os.Open(payload.Filepath)
		// Get the file from the form data
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

		// Encode the hash as a hexadecimal string
		hexHash := hex.EncodeToString(hash[:])

		// Create the destination file in the destination folder
		destinationFilePath := "files/" + hexHash
		destinationFile, err := os.Create(destinationFilePath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeStatusUpdate(w, "Cannot create the file to store base64 data.")
			return
		}
		defer destinationFile.Close()

		// Reset the read cursor back to the beginning of the file
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
func joinStrings(strings []string, delimiter string) string {
	if len(strings) == 0 {
		return ""
	}
	result := strings[0]
	for _, s := range strings[1:] {
		result += delimiter + s
	}
	return result
}
func getActivities(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		allActivities, err := backend.GetActivities()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeStatusUpdate(w, "Error getting the activities")
			return
		}
		var activityStrings []string
		for _, activity := range allActivities {
			activityString, err := json.Marshal(activity)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				writeStatusUpdate(w, "Failed to convert JSON Data into a string")
				return
			}
			activityStrings = append(activityStrings, string(activityString))
		}
		jsonArrayString := "[" + joinStrings(activityStrings, ",") + "]"
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonArrayString))

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only GET requests will be handled.")
		return
	}

}

func setActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			var payload Activity
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
				return
			}
			backend.SetActivity(payload)
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

type RemoveActivityJSONBody struct {
	Id int `json:"id"`
}

func removeActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			var payload RemoveActivityJSONBody
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
				return
			}
			backend.RemoveActivity(payload.Id)
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

type UpdateActivityJSONBody struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func updateActivityName(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			var payload UpdateActivityJSONBody
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
				return
			}
			backend.UpdateActivityName(payload.Id, payload.Name)
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

func InitAPIServer() {
	backend = NewBackend()
	peers = NewPeerStorage()
	fmt.Println("Settig up API Routes")
	publicKey, privateKey = orcaHash.LoadInKeys()
	orcaJobs.InitJobRoutes()
	http.HandleFunc("/file/", handleFileRoute)
	http.HandleFunc("/upload", uploadFile)

	http.HandleFunc("/getAllStored", getAllStored)
	http.HandleFunc("/get-file-info", getFileInfo)
	http.HandleFunc("/updateActivityName", updateActivityName)
	http.HandleFunc("/removeActivity", removeActivity)
	http.HandleFunc("/setActivity", setActivity)
	http.HandleFunc("/getActivities", getActivities)
	http.HandleFunc("/writeFile", writeFile)
	http.HandleFunc("/sendMoney", sendMoney)
	http.HandleFunc("/getLocation", getLocation)
	http.HandleFunc("/job-peer", JobPeerHandler)
}
