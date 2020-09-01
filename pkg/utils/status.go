package utils

const (
	StatusWaiting  int64 = 0
	StatusRunning  int64 = 1
	StatusFailed   int64 = 2
	StatusFinished int64 = 3
)

func GetStatus(status int64) string {
	if status == StatusWaiting {
		return "Waiting"
	}
	if status == StatusRunning {
		return "Running"
	}
	if status == StatusFailed {
		return "Failed"
	}
	if status == StatusFinished {
		return "Finished"
	}
	return "Unknown"
}
