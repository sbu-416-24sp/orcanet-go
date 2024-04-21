package jobs

import (
	"encoding/json"
	"errors"
	"os"
)

/*

### Endpoints
/add-job
  Description: Adds a job with an associated peer.
  PUT
  Parameters:
    fileHash: string
    peerID: PeerID
  Returns:
    {
      jobID: JobID
    }
    Success or error

/find-peer
  Description: Finds all peers there at are hosting a given file.
  GET
  Parameters:
    fileHash: string
  Returns:
    {
      peers: {
        peerID: PeerID
        ip: string
        region: string
        price: float
        reputation: int
      }[]
    }

/add-peer
  Description: Adds a peer to a job.
  PUT
  Parameters:
    jobID: JobID
    peerID: PeerID
  Returns:
    Success or error
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

/job-info
  Description: Gets the job information for a specific job
  GET
  Parameters:
    jobID: JobID
  Returns:
    {
      fileHash: string,
      fileName: string,
      fileSize: float,
      accumulatedMemory: float,
      accumulatedCost: float,
      projectedCost: float,
      eta: int,
      peer: PeerID,
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
/start-jobs
  Description: Begin using peer to download in job(s)
  PATCH
  Parameters:
    jobIDs: JobID[]
  Return:
    Success or error

/pause-jobs
  Description: Stop using peer to download in job(s)
  PATCH
  Parameters:
    jobIDs: JobID[]
  Return:
    Success or error

/terminate-jobs
  Description: Cancel job(s)
  PATCH
  Parameters:
    jobIDs: JobID[]
  Return:
    Success or error

/favorite-peer
  Description: Add peer to user’s favorite list
  PATCH
  Parameters:
    peerID: PeerID
  Return:
    Success or error

/unfavorite-peer
  Description: Delete peer in user’s favorite list
  PATCH
  Parameters:
    peerID: PeerID
  Return:
    Success or error

/like-peer
  Description: Adds a point to peer’s reputation
  PATCH
  Parameters:
    peerID: PeerID
  Return:
    Success or error

/unlike-peer
  Description: Removes a point in peer’s reputation
  PATCH
  Parameters:
    peerID: PeerID
  Return:
    Success or error

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
