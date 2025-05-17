package task

type Task struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
	Retries int                    `json:"retries"`
}

type Metrics struct {
	Total_jobs_in_queue int64 `json:"total_jobs_in_queue"`
	Jobs_done           int   `json:"jobs_done"`
	Jobs_failed         int   `json:"jobs_failed"`
}
