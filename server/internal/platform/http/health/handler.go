package health

import (
	"context"
	"encoding/json"
	"net/http"
)

type ReadinessChecker interface {
	Check(ctx context.Context) error
}

func LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(LivenessResponse{Status: "ok"})
	}
}

func ReadinessHandler(checker ReadinessChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
			return
		}

		checkResult := "up"
		status := http.StatusOK
		responseStatus := "ready"

		if err := checker.Check(r.Context()); err != nil {
			checkResult = "down"
			status = http.StatusServiceUnavailable
			responseStatus = "not_ready"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(ReadinessResponse{
			Status: responseStatus,
			Checks: ReadinessCheck{
				Database: checkResult,
			},
		})
	}
}
