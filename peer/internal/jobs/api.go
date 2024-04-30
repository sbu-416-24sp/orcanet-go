package jobs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
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

type FileChunkRequest struct {
	FileHash        string `json:"fileHash"`
	ChunkIndex           int `json:"chunkIndex"`
	JobId 				string `json:"jobId"`
}

type FileChunk struct {
	FileHash        string `json:"fileHash"`
	ChunkIndex           int `json:"chunkIndex"`
	MaxChunk           int `json:"maxChunk"`
	JobId              string `json:"jobId"`
	Data               []byte `json:"data"`
}

var Manager JobManager

func InitPeriodicJobSave() {
	Manager = JobManager{
		Jobs:    make([]Job, 0), // Initialize an empty slice of jobs
		Mutex:   sync.Mutex{},   // Initialize a mutex
		Changed: false,
	}
	for {
		time.Sleep(10 * time.Second)
		Manager.Mutex.Lock()
		if Manager.Changed {
			SaveHistory(Manager.Jobs)
		}
		Manager.Mutex.Unlock()
	}
}
func UpdateJobStatus(jobId string, status string) error {
	Manager.Mutex.Lock()
	for idx, job := range Manager.Jobs {
		if job.JobId == jobId {
			Manager.Jobs[idx].Status = status
			Manager.Changed = true
			break
		}
	}
	Manager.Mutex.Unlock()
	return nil
}
func GetJobStatus(jobId string) string {
	for _, job := range Manager.Jobs {
		if job.JobId == jobId {
			return job.Status
		}
	}
	return ""
}
func UpdateJobCost(jobId string, additionalCost int) error {
	Manager.Mutex.Lock()
	for idx, job := range Manager.Jobs {
		if job.JobId == jobId {
			Manager.Jobs[idx].AccumulatedCost += additionalCost
			Manager.Changed = true
			break
		}
	}
	Manager.Mutex.Unlock()
	return nil
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only PATCH requests will be handled.")
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeStatusUpdate(w, "Only GET requests will be handled.")
		return
	}
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
func JobListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		currentJobs := Manager.Jobs
		jsonData, err := json.Marshal(currentJobs)
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
		writeStatusUpdate(w, "Only GET requests will be handled.")
		return
	}
}

func GetHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		histories, err := LoadHistory()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeStatusUpdate(w, "Unable to read all histories")
		}
		jsonData, err := json.Marshal(histories)
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
		writeStatusUpdate(w, "Only GET requests will be handled.")
		return
	}
}

func InitJobRoutes() {
	http.HandleFunc("/terminate-jobs", TerminateJobsHandler)
	http.HandleFunc("/pause-jobs", PauseJobsHandler)
	http.HandleFunc("/job-info", JobInfoHandler)
	http.HandleFunc("/start-jobs", StartJobsHandler)
	http.HandleFunc("/remove-from-history", RemoveFromHistoryHandler)
	http.HandleFunc("/clear-history", ClearHistoryHandler)
	http.HandleFunc("/get-history", GetHistoryHandler)
	http.HandleFunc("/job-list", JobListHandler)
}
