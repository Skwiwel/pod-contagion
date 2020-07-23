package health

import (
	"net/http"
	"sync"
)

// Manager handles the Kubernetes health status changes of an application
type Manager interface {
	LivenessStatus() int
	ReadinessStatus() int
	SetLivenessStatus(status int) int
	SetReadinessStatus(status int) int
	LivenessHandler(w http.ResponseWriter, r *http.Request)
	ReadinessHandler(w http.ResponseWriter, r *http.Request)
}

type healthStatus struct {
	livenessStatus  int
	readinessStatus int
	mu              sync.RWMutex
}

// MakeHealthManager is a constructor for HealthManager
func MakeHealthManager() Manager {
	hs := healthStatus{
		livenessStatus:  http.StatusOK,
		readinessStatus: http.StatusOK,
	}
	return &hs
}

// LivenessStatus returns livenessStatus
func (hs *healthStatus) LivenessStatus() int {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.livenessStatus
}

// ReadinessStatus returns readinessStatus
func (hs *healthStatus) ReadinessStatus() int {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.readinessStatus
}

// SetLivenessStatus sets livenessStatus and returns the previous status
func (hs *healthStatus) SetLivenessStatus(status int) int {
	hs.mu.Lock()
	prevStatus := hs.livenessStatus
	hs.livenessStatus = status
	defer hs.mu.Unlock()
	return prevStatus
}

// SetReadinessStatus sets readinessStatus and returns the previous status
func (hs *healthStatus) SetReadinessStatus(status int) int {
	hs.mu.Lock()
	prevStatus := hs.readinessStatus
	hs.readinessStatus = status
	defer hs.mu.Unlock()
	return prevStatus
}

// LivenessHandler responds to health check requests.
func (hs *healthStatus) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(hs.LivenessStatus())
}

// ReadinessHandler responds to readiness check requests.
func (hs *healthStatus) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(hs.ReadinessStatus())
}
