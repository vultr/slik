package slurm

const (
	WorkloadStatusPending   string = "Pending"
	WorkloadStatusRunning   string = "Running"
	WorkloadStatusCompleted string = "Completed"
	WorkloadStatusFailed    string = "Failed"
	WorkloadStatusUnknown   string = "Unknown"
)

const (
	ConflictRetryIntervalSec int64 = 1
)
