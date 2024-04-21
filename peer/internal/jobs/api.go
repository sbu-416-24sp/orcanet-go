package jobs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	FileHash        string `json:"fileHash"`
	JobId           string `json:"jobID"`
	TimeQueued      string `json:"timeQueued"`
	Status          string `json:"status"`
	AccumulatedCost int    `json:"accumulatedCost"`
	ProjectedCost   int    `json:"projectedCost"`
	ETA             int    `json:"eta"`
	PeerId          string `json:"peer"`
}

type JobManager struct {
	Jobs    []Job
	Mutex   sync.Mutex
	Changed bool
}

var manager JobManager

func InitPeriodicJobSave() {
	manager = JobManager{
		Jobs:    make([]Job, 0), // Initialize an empty slice of jobs
		Mutex:   sync.Mutex{},   // Initialize a mutex
		Changed: false,
	}
	for {
		time.Sleep(10 * time.Second)
		manager.Mutex.Lock()
		if manager.Changed {
			SaveHistory(manager.Jobs)
		}
		manager.Mutex.Unlock()
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

type RmFromHistoryReqPayload struct {
	JobId string `json:"jobID"`
}

func RemoveFromHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPatch {
		var payload RmFromHistoryReqPayload
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
			return
		}
		err := RemoveFromHistory(payload.JobId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		writeStatusUpdate(w, "success")
		return
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only PATCH requests will be handled.")
		return
	}
}

func ClearHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPatch {
		ClearHistory()
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only PATCH requests will be handled.")
		return
	}
}

type AddJobReqPayload struct {
	FileHash string `json:"fileHash"`
	PeerId   string `json:"peer"`
}

type AddJobResPayload struct {
	JobId string `json:"jobID"`
}

func AddJobHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		var payload AddJobReqPayload
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
			return
		}
		id := uuid.New()
		timeString := time.Now().Format(time.RFC3339)
		newJob := Job{
			FileHash:        payload.FileHash,
			JobId:           id.String(),
			TimeQueued:      timeString,
			Status:          "paused",
			AccumulatedCost: 0,
			ProjectedCost:   -1,
			ETA:             -1,
			PeerId:          payload.PeerId,
		}
		AddJob(newJob)
		w.WriteHeader(http.StatusOK)
		writeStatusUpdate(w, "Successfully added job.")
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only PUT requests will be handled.")
		return
	}
}

type JobInfoReqPayload struct {
	JobId string `json:"jobID"`
}

func JobInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		queryParams := r.URL.Query()
		jobId := queryParams.Get("jobID")
		job, err := FindJob(jobId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, err.Error())
			return
		}
		jsonData, err := json.Marshal(job)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeStatusUpdate(w, "Failed to convert JSON Data into a string")
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only PUT requests will be handled.")
		return
	}
}

type JobPeerReqPayload struct {
	FileHash string `json:"fileHash"`
	PeerId   string `json:"peer"`
}

type JobPeerResPayload struct {
	JobId             string `json:"ipAddress"`
	Region            string `json:"region"`
	Liked             bool   `json:"liked"`
	Status            string `json:"status"`
	AccumulatedMemory string `json:"accumulatedMemory"`
	Price             string `json:"price"`
}

func JobPeerHandler() {

}
func StartJobsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPatch {
		var jobIds []JobInfoReqPayload
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&jobIds); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
			return
		}
		for _, jobId := range jobIds {
			err := StartJob(jobId.JobId)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, err.Error())
			}
		}
		w.WriteHeader(http.StatusOK)
		writeStatusUpdate(w, "Successfully added job.")
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only PATCH requests will be handled.")
	}
}

func PauseJobsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPatch {
		var jobIds []JobInfoReqPayload
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&jobIds); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
			return
		}
		for _, jobId := range jobIds {
			err := PauseJob(jobId.JobId)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, err.Error())
			}
		}
		w.WriteHeader(http.StatusOK)
		writeStatusUpdate(w, "Successfully added job.")
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only PATCH requests will be handled.")
	}
}

func TerminateJobsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPatch {
		var jobIds []JobInfoReqPayload
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&jobIds); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writeStatusUpdate(w, "Cannot marshal payload in Go object. Does the payload have the correct body structure?")
			return
		}
		for _, jobId := range jobIds {
			err := TerminateJob(jobId.JobId)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeStatusUpdate(w, err.Error())
			}
		}
		w.WriteHeader(http.StatusOK)
		writeStatusUpdate(w, "Successfully added job.")
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only PATCH requests will be handled.")
		return
	}
}

func InitJobRoutes() {
	http.HandleFunc("/terminate-jobs", TerminateJobsHandler)
	http.HandleFunc("/pause-jobs", PauseJobsHandler)
	http.HandleFunc("/job-info", JobInfoHandler)
	http.HandleFunc("/add-job", AddJobHandler)
	http.HandleFunc("/remove-from-history", RemoveFromHistoryHandler)
	http.HandleFunc("/clear-history", ClearHistoryHandler)
}
