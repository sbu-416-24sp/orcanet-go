package jobs

import (
	"encoding/json"
	"errors"
	"os"
)

/*

### Endpoints

/job-list
  Description: Returns list of current jobs.
  GET
  Parameters:
    none
  Returns:
    {
      jobs: {
        jobID: JobID,
        fileName: string,
        fileSize: int // bytes,
        eta: int // seconds
        timeQueued: Date,
        status: "active" | "paused" | "error" | "completed",
      }[]
*/

func AddJob(job Job) {
	Manager.Mutex.Lock()
	Manager.Jobs = append(Manager.Jobs, job)
	Manager.Changed = true
	Manager.Mutex.Unlock()
}

func LoadHistory() ([]Job, error) {
	fileData, err := os.ReadFile("./internal/jobs/jobs.json")
	if err != nil {
		return nil, err
	}
	var jobs []Job
	err = json.Unmarshal(fileData, &jobs)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}
func SaveHistory(jobs []Job) error {
	Manager.Changed = false
	jsonData, err := json.Marshal(jobs)
	if err != nil {
		return err
	}
	err = os.WriteFile("./internal/jobs/jobs.json", jsonData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func RemoveFromHistory(jobId string) error {
	Manager.Mutex.Lock()
	for idx, job := range Manager.Jobs {
		if job.JobId == jobId {
			Manager.Jobs = append(Manager.Jobs[:idx], Manager.Jobs[idx+1:]...)
			Manager.Changed = true
			Manager.Mutex.Unlock()
			return nil
		}
	}
	Manager.Mutex.Unlock()
	return errors.New("unable to find job that matches jobID")
}

func ClearHistory() {
	Manager.Mutex.Lock()
	newJobs := make([]Job, 0)
	for _, job := range Manager.Jobs {
		if job.Status != "completed" {
			Manager.Changed = true
			newJobs = append(newJobs, job)
		}
	}
	Manager.Jobs = newJobs
	Manager.Mutex.Unlock()
}

func TerminateJob(jobId string) error {
	Manager.Mutex.Lock()
	for idx, job := range Manager.Jobs {
		if job.JobId == jobId {
			Manager.Jobs[idx].Status = "terminated"
			Manager.Changed = true
			Manager.Mutex.Unlock()
			return nil
		}
	}
	Manager.Mutex.Unlock()
	return errors.New("Unable to find jobId: " + jobId)
}

func PauseJob(jobId string) error {
	Manager.Mutex.Lock()
	for idx, job := range Manager.Jobs {
		if job.JobId == jobId {
			Manager.Jobs[idx].Status = "paused"
			Manager.Changed = true
			Manager.Mutex.Unlock()
			return nil
		}
	}
	Manager.Mutex.Unlock()
	return errors.New("Unable to find jobId: " + jobId)
}

func StartJob(jobId string) error {
	Manager.Mutex.Lock()
	for idx, job := range Manager.Jobs {
		if job.JobId == jobId {
			Manager.Jobs[idx].Status = "active"
			Manager.Changed = true
			Manager.Mutex.Unlock()
			return nil
		}
	}
	Manager.Mutex.Unlock()
	return errors.New("Unable to find jobId: " + jobId)
}

func FindJob(jobId string) (Job, error) {
	Manager.Mutex.Lock()
	for _, job := range Manager.Jobs {
		if job.JobId == jobId {
			Manager.Mutex.Unlock()
			return job, nil
		}
	}
	Manager.Mutex.Unlock()
	return Job{}, errors.New("unable to find job with specified jobId")
}
