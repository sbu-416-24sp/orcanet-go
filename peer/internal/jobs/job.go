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
    }


/job-peer
  Description: Gets the peer information for a specific job
  GET
  Parameters:
    jobID: JobID
    peerID: PeerID
  Returns:
    {
      ipAddress: string,
      region: string,
      liked: boolean,
      status: "uninitialized" | "active" | "paused" | "error" | "terminated",
      accumulatedMemory: float,
      price: int,
      graph: { time: float, speed: float }
    }



*/

func AddJob(job Job) {
	manager.Mutex.Lock()
	manager.Jobs = append(manager.Jobs, job)
	manager.Changed = true
	manager.Mutex.Unlock()
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
	manager.Changed = false
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
	manager.Mutex.Lock()
	for idx, job := range manager.Jobs {
		if job.JobId == jobId {
			manager.Jobs = append(manager.Jobs[:idx], manager.Jobs[idx+1:]...)
			manager.Changed = true
			manager.Mutex.Unlock()
			return nil
		}
	}
	manager.Mutex.Unlock()
	return errors.New("unable to find job that matches jobID")
}

func ClearHistory() {
	manager.Mutex.Lock()
	newJobs := make([]Job, 0)
	for _, job := range manager.Jobs {
		if job.Status != "completed" {
			manager.Changed = true
			newJobs = append(newJobs, job)
		}
	}
	manager.Jobs = newJobs
	manager.Mutex.Unlock()
}

func TerminateJob(jobId string) error {
	manager.Mutex.Lock()
	for idx, job := range manager.Jobs {
		if job.JobId == jobId {
			manager.Jobs[idx].Status = "terminated"
			manager.Changed = true
			manager.Mutex.Unlock()
			return nil
		}
	}
	manager.Mutex.Unlock()
	return errors.New("Unable to find jobId: " + jobId)
}

func PauseJob(jobId string) error {
	manager.Mutex.Lock()
	for idx, job := range manager.Jobs {
		if job.JobId == jobId {
			manager.Jobs[idx].Status = "paused"
			manager.Changed = true
			manager.Mutex.Unlock()
			return nil
		}
	}
	manager.Mutex.Unlock()
	return errors.New("Unable to find jobId: " + jobId)
}

func StartJob(jobId string) error {
	manager.Mutex.Lock()
	for idx, job := range manager.Jobs {
		if job.JobId == jobId {
			manager.Jobs[idx].Status = "active"
			manager.Changed = true
			manager.Mutex.Unlock()
			return nil
		}
	}
	manager.Mutex.Unlock()
	return errors.New("Unable to find jobId: " + jobId)
}

func FindJob(jobId string) (Job, error) {
	manager.Mutex.Lock()
	for _, job := range manager.Jobs {
		if job.JobId == jobId {
			manager.Mutex.Unlock()
			return job, nil
		}
	}
	manager.Mutex.Unlock()
	return Job{}, errors.New("unable to find job with specified jobId")
}
