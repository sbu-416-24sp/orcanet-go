package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type PeerIdPOSTPayload struct {
	PeerID string `json:"peerID"`
}

func getAllPeers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		peerTable := GetPeerTable()
		var peers []PeerInfo
		for _, peer := range peerTable {
			peers = append(peers, peer)
		}
		jsonPeers, err := json.Marshal(peers)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeStatusUpdate(w, "Failed to convert all Peer JSON Data into a string")
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonPeers)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only GET requests will be handled.")
		return
	}
}
func getPeer(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		queryParams := r.URL.Query()
		peerId := queryParams.Get("peer-id")
		peerTable := GetPeerTable()
		if peer, ok := peerTable[peerId]; ok {
			jsonPeer, err := json.Marshal(peer)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				writeStatusUpdate(w, "Failed to convert Peer JSON Data into a string")
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			w.Write(jsonPeer)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Peer ID not found inside list of connected peer ids.")
			return
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only GET requests will be handled.")
		return
	}
}

func removePeer(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		peerTable := GetPeerTable()
		if peer, ok := peerTable[peerId]; ok {
			jsonPeer, err := json.Marshal(peer)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				writeStatusUpdate(w, "Failed to convert Peer JSON Data into a string")
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			w.Write(jsonPeer)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Peer ID not found inside list of connected peer ids.")
			return
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only POST requests will be handled.")
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
