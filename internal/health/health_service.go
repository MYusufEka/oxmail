package health

import (
	"context"
	"sync"
	"time"
)

// ServiceChecker defines the interface for checking a service's health.
type ServiceChecker interface {
	Check(ctx context.Context) (healthy bool, latencyMs int)
}

// ServiceResult represents the health check result for a single service.
type ServiceResult struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	LatencyMs int    `json:"latencyMs"`
}

// Result is the aggregated health response.
type Result struct {
	Status   string          `json:"status"`
	Version  string          `json:"version"`
	Uptime   string          `json:"uptime"`
	Services []ServiceResult `json:"services"`
}

// coreServices are services whose failure marks the system unhealthy.
var coreServices = map[string]bool{
	"postfix": true,
	"dovecot": true,
}

// Service aggregates health checks across all services.
type Service struct {
	checkers  map[string]ServiceChecker
	version   string
	startTime time.Time

	mu       sync.Mutex
	cached   *Result
	cachedAt time.Time
	cacheTTL time.Duration
}

// NewService creates a Service with the given checkers and version.
func NewService(checkers map[string]ServiceChecker, version string) *Service {
	return &Service{
		checkers:  checkers,
		version:   version,
		startTime: time.Now(),
		cacheTTL:  5 * time.Second,
	}
}

// Check performs health checks on all services and returns an aggregated result.
// Results are cached for 5 seconds.
func (h *Service) Check(ctx context.Context) Result {
	h.mu.Lock()
	if h.cached != nil && time.Since(h.cachedAt) < h.cacheTTL {
		result := *h.cached
		h.mu.Unlock()
		return result
	}
	h.mu.Unlock()

	services := make([]ServiceResult, 0, len(h.checkers))

	for name, checker := range h.checkers {
		healthy, latencyMs := checker.Check(ctx)
		status := "healthy"
		if !healthy {
			status = "unhealthy"
		}
		services = append(services, ServiceResult{
			Name:      name,
			Status:    status,
			LatencyMs: latencyMs,
		})
	}

	aggregatedStatus := aggregateStatus(services)
	uptime := time.Since(h.startTime).Round(time.Second).String()

	result := Result{
		Status:   aggregatedStatus,
		Version:  h.version,
		Uptime:   uptime,
		Services: services,
	}

	h.mu.Lock()
	h.cached = &result
	h.cachedAt = time.Now()
	h.mu.Unlock()

	return result
}

// ExpireCache forces the cache to expire (used in tests).
func (h *Service) ExpireCache() {
	h.mu.Lock()
	h.cached = nil
	h.mu.Unlock()
}

// SetStartTime overrides the start time (used in tests).
func (h *Service) SetStartTime(t time.Time) {
	h.startTime = t
}

// aggregateStatus determines overall status from individual service results.
func aggregateStatus(services []ServiceResult) string {
	coreDown := false
	anyDown := false

	for _, svc := range services {
		if svc.Status == "unhealthy" {
			anyDown = true
			if coreServices[svc.Name] {
				coreDown = true
			}
		}
	}

	if coreDown {
		return "unhealthy"
	}
	if anyDown {
		return "degraded"
	}
	return "healthy"
}
