package api

import (
	"crypto/sha256"
	"encoding/json"
	"net/http"
	orcaClient "orca-peer/internal/client"
	orcaStatus "orca-peer/internal/status"
	"os"
)

// API to use with out CLI

/*
Location
SendMoney
Network
*/
type SendMoneyJSONRequest struct {
	Amount     float64 `json:"amount"`
	ServerIp   string  `json:"host"`
	ServerPort string  `json:"port"`
}

func sendMoney(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			// For JSON content type, decode the JSON into a struct
			var payload SendMoneyJSONRequest
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
				return
			}
			orcaClient.SendTransaction(payload.Amount, payload.ServerIp, payload.ServerPort, publicKey, privateKey)
			return
		default:
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Request must have the content header set as application/json")
			return
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only POST requests will be handled")
		return
	}
}

type LocationInfoResponse struct {
	Ip        string `json:"ip"`
	Network   string `json:"network"`
	City      string `json:"city"`
	Region    string `json:"region"`
	Country   string `json:"country"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	ASN       string `json:"asn"`
	Timezone  string `json:"timezone"`
	Continent string `json:"continent"`
	Org       string `json:"org`
}

func getLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		locationJsonString := orcaStatus.GetLocationData()
		var locationJson map[string]interface{}
		err := json.Unmarshal([]byte(locationJsonString), &locationJson)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Issue with reading the request body. Is the correct format used?")
			return
		}
		location := LocationInfoResponse{}
		location.Ip = locationJson["ip"].(string)
		location.Network = locationJson["network"].(string)
		location.City = locationJson["city"].(string)
		location.Region = locationJson["region"].(string)
		location.Country = locationJson["country_name"].(string)
		location.Latitude = locationJson["latitude"].(string)
		location.Longitude = locationJson["longitude"].(string)
		location.ASN = locationJson["asn"].(string)
		location.Timezone = locationJson["timezone"].(string)
		location.Continent = locationJson["continent_code"].(string)
		location.Org = locationJson["org"].(string)

		jsonData, err := json.Marshal(location)
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
		writeStatusUpdate(w, "Only GET requests will be handled")
		return
	}
}

func hashFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			// For JSON content type, decode the JSON into a struct
			var payload UploadFileJSONBody
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
				return
			}
			fileData, err := os.ReadFile("files/" + payload.Filepath)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				writeStatusUpdate(w, "Failed to read in file from given path")
				return
			}
			hash := sha256.Sum256(fileData)
			responseMsg := map[string]interface{}{
				"hash": hash,
			}
			responseMsgJsonString, err := json.Marshal(responseMsg)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				writeStatusUpdate(w, "Unable to send hash back to user inside JSON object, issue when querying.")
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write(responseMsgJsonString)
			return
		default:
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Request must have the content header set as application/json")
			return
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only GET requests will be handled")
		return
	}
}
