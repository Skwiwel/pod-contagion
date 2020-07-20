package health

import (
	"net/http"
	"sync"
)

var (
	livenessStatus  = http.StatusOK
	readinessStatus = http.StatusOK
	mu              sync.RWMutex
)

// LivenessStatus returns livenessStatus
func LivenessStatus() int {
	mu.RLock()
	defer mu.RUnlock()
	return livenessStatus
}

// ReadinessStatus returns readinessStatus
func ReadinessStatus() int {
	mu.RLock()
	defer mu.RUnlock()
	return readinessStatus
}

// SetLivenessStatus sets livenessStatus
func SetLivenessStatus(status int) {
	mu.Lock()
	livenessStatus = status
	mu.Unlock()
}

// SetReadinessStatus sets readinessStatus
func SetReadinessStatus(status int) {
	mu.Lock()
	readinessStatus = status
	mu.Unlock()
}

// LivenessHandler responds to health check requests.
func LivenessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(LivenessStatus())
}

// ReadinessHandler responds to readiness check requests.
func ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(ReadinessStatus())
}

// LivenessStatusHandler toggles livenessStatus
func LivenessStatusHandler(w http.ResponseWriter, r *http.Request) {
	switch LivenessStatus() {
	case http.StatusOK:
		SetLivenessStatus(http.StatusServiceUnavailable)
	case http.StatusServiceUnavailable:
		SetLivenessStatus(http.StatusOK)
	}
	w.WriteHeader(http.StatusOK)
}

// ReadinessStatusHandler toggles readinessStatus
func ReadinessStatusHandler(w http.ResponseWriter, r *http.Request) {
	switch ReadinessStatus() {
	case http.StatusOK:
		SetReadinessStatus(http.StatusServiceUnavailable)
	case http.StatusServiceUnavailable:
		SetReadinessStatus(http.StatusOK)
	}
	w.WriteHeader(http.StatusOK)
}
