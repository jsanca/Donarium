package health

type LivenessResponse struct {
	Status string `json:"status"`
}

type ReadinessCheck struct {
	Database string `json:"database"`
}

type ReadinessResponse struct {
	Status string         `json:"status"`
	Checks ReadinessCheck `json:"checks"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
