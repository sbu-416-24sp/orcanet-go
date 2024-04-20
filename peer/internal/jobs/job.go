package job

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
type JobInfo struct {
}

func LoadHistory() {

}
func SaveHistory() {

}
func RemoveFromHistoryHandler() {

}

func ClearHistoryHandler() {

}

type AddJobReqPayload struct {
	FileHash string `json:"fileHash"`
	PeerId   string `json:"peer"`
}

type AddJobResPayload struct {
	JobId string `json:"jobID"`
}

func AddJobHandler() {

}

type JobInfoReqPayload struct {
	JobId string `json:"jobID"`
}

func JobInfoHandler() {

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
func StartJobsHandler() {

}

func PauseJobsHandler() {

}

func TerminateJobsHandler() {

}
