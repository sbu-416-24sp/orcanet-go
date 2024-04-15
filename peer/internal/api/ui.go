package api

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	orcaHash "orca-peer/internal/hash"
	"os"
	"path/filepath"
)

type GetFileJSONBody struct {
	Filename string `json:"filename"`
	CID      string `json:"cid"`
}

type UploadFileJSONBody struct {
	Filepath string `json:"filepath"`
}

var backend *Backend
var peers *PeerStorage
var publicKey *rsa.PublicKey
var privateKey *rsa.PrivateKey

func getFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
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
			if payload.Filename == "" && payload.CID == "" {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Missing CID and Filename field in request")
				return
			}
			fileaddress := ""

			if _, err := os.Stat("files/stored/" + payload.Filename); !os.IsNotExist(err) {
				fileaddress = "files/stored/" + payload.Filename
			}
			if _, err := os.Stat("files/requested/" + payload.Filename); !os.IsNotExist(err) && fileaddress == "" {
				fileaddress = "files/requested/" + payload.Filename
			}
			if _, err := os.Stat("files/" + payload.Filename); !os.IsNotExist(err) && fileaddress == "" {
				fileaddress = "files/" + payload.Filename
			}
			if fileaddress != "" {

			} else {
				w.WriteHeader(http.StatusInternalServerError)
				writeStatusUpdate(w, "Cannot find specified file inside files directory")
				return
			}
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

type FileInfo struct {
	Filename     string `json:"filename"`
	Filesize     int    `json:"filesize"`
	Filehash     string `json:"filehash"`
	Lastmodified string `json:"lastmodified"`
	Filecontent  string `json:"filecontent"`
}

func getFileInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		queryParams := r.URL.Query()

		// Retrieve specific query parameters by key
		filename := queryParams.Get("filename")
		if st, err := os.Stat("files/" + filename); !os.IsNotExist(err) {
			fileData, err := os.ReadFile("files/" + filename)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				writeStatusUpdate(w, "Failed to read in file from given path")
				return
			}
			lenData := len(fileData)
			base64Encode := base64.StdEncoding.EncodeToString(fileData)
			hash := sha256.Sum256(fileData)

			// Encode the hash as a hexadecimal string
			hexHash := hex.EncodeToString(hash[:])

			fileInfoResp := FileInfo{
				Filename:     filename,
				Filesize:     lenData,
				Filehash:     hexHash,
				Lastmodified: st.ModTime().String(),
				Filecontent:  base64Encode,
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

func uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			var payload UploadFileJSONBody
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
				return
			}
			fileData, err := os.ReadFile(payload.Filepath)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				writeStatusUpdate(w, "Error reading in file from the Filepath specified.")
				return
			}
			if _, err := os.Stat(payload.Filepath); !os.IsNotExist(err) {
				sourceFile, err := os.Open(payload.Filepath)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					writeStatusUpdate(w, "Error getting information about file from file system.")
					return
				}
				defer sourceFile.Close()
				hash := sha256.Sum256(fileData)

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

				_, err = io.Copy(destinationFile, sourceFile)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					writeStatusUpdate(w, "Unable to copy base64 data.")
					return
				}
				w.WriteHeader(http.StatusOK)
				writeStatusUpdate(w, "Successfully uploaded file from local computer into files directory")
				return
			} else {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "File specified does not exist.")
				return
			}

		default:
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Successfully uploaded file from local computer into files directory")
			return
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only POST requests will be handled.")
		return
	}
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
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
			if payload.Filename == "" && payload.CID == "" {
				w.WriteHeader(http.StatusInternalServerError)
				writeStatusUpdate(w, "Missing Filename and CID values inside of the payload.")
				return
			}
			fileDir := "./files"
			var filePath string

			// Check if the file exists in the "stored" directory
			storedFilePath := filepath.Join(fileDir, "stored", payload.Filename)
			if _, err := os.Stat(storedFilePath); err == nil {
				filePath = storedFilePath
			}
			// Check if the file exists in the "requested" directory
			requestedFilePath := filepath.Join(fileDir, "requested", payload.Filename)
			if _, err := os.Stat(requestedFilePath); err == nil {
				filePath = requestedFilePath
			}

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
		writeStatusUpdate(w, "Only POST requests will be handled.")
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

func addPeer(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			var payload PeerInfo
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
				return
			}
			peers.AddPeer(payload)
			return
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

type PeerIdPOSTPayload struct {
	PeerID string `json:"peerID"`
}

func getPeer(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			var payload PeerIdPOSTPayload
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
				return
			}
			currPeer, exists := peers.GetPeer(payload.PeerID)
			if exists {
				peerString, err := json.Marshal(currPeer)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					writeStatusUpdate(w, "Failed to marshal data from go object into a string")
					return
				}
				w.Header().Set("Content-Type", "application/octet-stream")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(peerString))
				return
			} else {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Unable to find a string with the given peer id")
				return
			}

		default:
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Request must have the content header set as application/json")
			return
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only GET requests will be handled.")
		return
	}

}
func getAllPeers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			var payload PeerIdPOSTPayload
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
				return
			}
			allPeers := peers.GetAllPeers()
			var peerStrings []string
			for _, peer := range allPeers {
				peerString, err := json.Marshal(peer)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					writeStatusUpdate(w, "Issue converting all peers into a string")
					return
				}
				peerStrings = append(peerStrings, string(peerString))
			}
			jsonArrayPeerString := "[" + joinStrings(peerStrings, ",") + "]"
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(jsonArrayPeerString))
			return
		default:
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Request must have the content header set as application/json")
			return
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only GET requests will be handled.")
		return
	}

}
func updatePeer(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			var payload PeerInfo
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
				return
			}
			peers.UpdatePeer(payload.PeerID, payload)
			return
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

func removePeer(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			var payload PeerIdPOSTPayload
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
				return
			}
			peers.RemovePeer(payload.PeerID)
			return
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

func InitServer() {
	backend = NewBackend()
	peers = NewPeerStorage()
	publicKey, privateKey = orcaHash.LoadInKeys()
	http.HandleFunc("/getFile", getFile)
	http.HandleFunc("/getFileInfo", getFileInfo)
	http.HandleFunc("/uploadFile", uploadFile)
	http.HandleFunc("/deleteFile", deleteFile)
	http.HandleFunc("/updateActivityName", updateActivityName)
	http.HandleFunc("/removeActivity", removeActivity)
	http.HandleFunc("/setActivity", setActivity)
	http.HandleFunc("/getActivities", getActivities)
	http.HandleFunc("/writeFile", writeFile)
	http.HandleFunc("/removePeer", removePeer)
	http.HandleFunc("/updatePeer", updatePeer)
	http.HandleFunc("/getAllPeers", getAllPeers)
	http.HandleFunc("/getPeer", getPeer)
	http.HandleFunc("/addPeer", addPeer)
	http.HandleFunc("/sendMoney", sendMoney)
	http.HandleFunc("/getLocation", getLocation)
}
