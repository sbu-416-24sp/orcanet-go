package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	orcaJob "orca-peer/internal/jobs"
	orcaServer "orca-peer/internal/server"
)

type JobPeerResPayload struct {
	IpAddress         string `json:"ipAddress"`
	Region            string `json:"region"`
	Liked             bool   `json:"liked"`
	Status            string `json:"status"`
	AccumulatedMemory string `json:"accumulatedMemory"`
	Price             string `json:"price"`
}

func JobPeerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		queryParams := r.URL.Query()
		filehash := queryParams.Get("fileHash")
		peerId := queryParams.Get("peer")
		peers := orcaServer.GetPeerTable()
		if val, ok := peers[peerId]; ok {
			var currJob orcaJob.Job
			found := false
			for _, job := range orcaJob.Manager.Jobs {
				if job.FileHash == filehash {
					currJob = job
					found = true
					break
				}
			}
			if !found {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, "Unable to find a job with specified job id")
				return
			}
			peerOnJob := JobPeerResPayload{
				IpAddress:         val.Connection,
				Region:            val.Location,
				Liked:             false,
				Status:            currJob.Status,
				AccumulatedMemory: fmt.Sprint(currJob.AccumulatedCost),
				Price:             fmt.Sprint(currJob.ProjectedCost),
			}
			jsonData, err := json.Marshal(peerOnJob)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				writeStatusUpdate(w, "Failed to convert JSON Data into a string")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(jsonData)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Peer with specified ID does not exist")
		}
		w.WriteHeader(http.StatusOK)
		writeStatusUpdate(w, "Successfully added job.")
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only PATCH requests will be handled.")
	}
}
