package mining

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type PutDeviceRequestBody struct {
	PublicKey string `json:"pub_key"`
	Switch    string `json:"switch"`
}

func PutDeviceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			var payload PutDeviceRequestBody
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

type GetDevicesRequestBody struct {
}

func GetDeviceHandler(w http.ResponseWriter, r *http.Request) {

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
