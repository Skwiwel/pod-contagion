package health

import (
	"net/http"
	"sync"
)

// Manager handles the Kubernetes health status changes of an application
type Manager interface {
	LivenessStatus() int
	ReadinessStatus() int
	SetLivenessStatus(status int)
	SetReadinessStatus(status int)
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

// SetLivenessStatus sets livenessStatus
func (hs *healthStatus) SetLivenessStatus(status int) {
	hs.mu.Lock()
	hs.livenessStatus = status
	hs.mu.Unlock()
}

// SetReadinessStatus sets readinessStatus
func (hs *healthStatus) SetReadinessStatus(status int) {
	hs.mu.Lock()
	hs.readinessStatus = status
	hs.mu.Unlock()
}

// LivenessHandler responds to health check requests.
func (hs *healthStatus) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(hs.LivenessStatus())
}

// ReadinessHandler responds to readiness check requests.
func (hs *healthStatus) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(hs.ReadinessStatus())
}
