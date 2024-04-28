package api

import (
	"encoding/json"
	"net/http"
)

type WalletPassphraseJSONRequest struct {
	Wallet_Address string  `json:"walletName"`
	Time           float64 `json:"timeUnlock"`
}

func walletPassphrase(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			// For JSON content type, decode the JSON into a struct
			var payload WalletPassphraseJSONRequest
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
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

type SendToAddressRequest struct {
	Wallet_Address string  `json:"walletAddress"`
	Amount         float64 `json:"amount"`
}
type SendToAddressResponse struct {
	Hash string `json:"hash"`
}

func sendToAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			var payload SendToAddressRequest
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
				return
			}
			// Check if wallet is unlocked

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

type GenerateJSONRequest struct {
	Blocks int `json:"blocks"`
}
type GenerateJSONResponse struct {
	Hash string `json:"hash"`
}

func generateCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			// For JSON content type, decode the JSON into a struct
			var payload GenerateJSONRequest
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
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

type GetBalanceJSONResponse struct {
	Balance int `json:"balance"`
}

func getBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		//Need to execute command
		fileInfoResp := GetBalanceJSONResponse{
			Balance: 0,
		}
		jsonData, err := json.Marshal(fileInfoResp)
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
		writeStatusUpdate(w, "Request must have the content header set as application/json")
		return
	}
}
