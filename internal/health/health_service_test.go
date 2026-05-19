package health

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockChecker implements ServiceChecker for testing.
type mockChecker struct {
	healthy   bool
	latencyMs int
}

func (m *mockChecker) Check(_ context.Context) (bool, int) {
	return m.healthy, m.latencyMs
}

// countingChecker tracks how many times Check is called.
type countingChecker struct {
	callCount *int
	healthy   bool
	latencyMs int
}

func (c *countingChecker) Check(_ context.Context) (bool, int) {
	*c.callCount++
	return c.healthy, c.latencyMs
}

func TestService_AllHealthy(t *testing.T) {
	checkers := map[string]ServiceChecker{
		"postfix": &mockChecker{healthy: true, latencyMs: 2},
		"dovecot": &mockChecker{healthy: true, latencyMs: 1},
		"rspamd":  &mockChecker{healthy: true, latencyMs: 5},
		"redis":   &mockChecker{healthy: true, latencyMs: 1},
	}

	svc := NewService(checkers, "0.1.0")
	result := svc.Check(context.Background())

	assert.Equal(t, "healthy", result.Status)
	assert.Equal(t, "0.1.0", result.Version)
	assert.NotEmpty(t, result.Uptime)
	require.Len(t, result.Services, 4)

	for _, s := range result.Services {
		assert.Equal(t, "healthy", s.Status)
	}
}

func TestService_DegradedWhenNonCoreDown(t *testing.T) {
	checkers := map[string]ServiceChecker{
		"postfix": &mockChecker{healthy: true, latencyMs: 2},
		"dovecot": &mockChecker{healthy: true, latencyMs: 1},
		"rspamd":  &mockChecker{healthy: false, latencyMs: 0},
		"redis":   &mockChecker{healthy: false, latencyMs: 0},
	}

	svc := NewService(checkers, "0.1.0")
	result := svc.Check(context.Background())

	assert.Equal(t, "degraded", result.Status)

	serviceMap := serviceSliceToMap(result.Services)
	assert.Equal(t, "healthy", serviceMap["postfix"].Status)
	assert.Equal(t, "healthy", serviceMap["dovecot"].Status)
	assert.Equal(t, "unhealthy", serviceMap["rspamd"].Status)
	assert.Equal(t, "unhealthy", serviceMap["redis"].Status)
}

func TestService_UnhealthyWhenPostfixDown(t *testing.T) {
	checkers := map[string]ServiceChecker{
		"postfix": &mockChecker{healthy: false, latencyMs: 0},
		"dovecot": &mockChecker{healthy: true, latencyMs: 1},
		"rspamd":  &mockChecker{healthy: true, latencyMs: 5},
		"redis":   &mockChecker{healthy: true, latencyMs: 1},
	}

	svc := NewService(checkers, "0.1.0")
	result := svc.Check(context.Background())

	assert.Equal(t, "unhealthy", result.Status)
}

func TestService_UnhealthyWhenDovecotDown(t *testing.T) {
	checkers := map[string]ServiceChecker{
		"postfix": &mockChecker{healthy: true, latencyMs: 2},
		"dovecot": &mockChecker{healthy: false, latencyMs: 0},
		"rspamd":  &mockChecker{healthy: true, latencyMs: 5},
		"redis":   &mockChecker{healthy: true, latencyMs: 1},
	}

	svc := NewService(checkers, "0.1.0")
	result := svc.Check(context.Background())

	assert.Equal(t, "unhealthy", result.Status)
}

func TestService_CachesResultsFor5Seconds(t *testing.T) {
	callCount := 0
	checkers := map[string]ServiceChecker{
		"postfix": &countingChecker{callCount: &callCount, healthy: true, latencyMs: 2},
		"dovecot": &mockChecker{healthy: true, latencyMs: 1},
		"rspamd":  &mockChecker{healthy: true, latencyMs: 5},
		"redis":   &mockChecker{healthy: true, latencyMs: 1},
	}

	svc := NewService(checkers, "0.1.0")

	// First call performs checks
	result1 := svc.Check(context.Background())
	assert.Equal(t, "healthy", result1.Status)
	assert.Equal(t, 1, callCount)

	// Second call within 5s returns cached result
	result2 := svc.Check(context.Background())
	assert.Equal(t, "healthy", result2.Status)
	assert.Equal(t, 1, callCount)

	// Expire the cache manually
	svc.ExpireCache()

	// Third call after expiry performs checks again
	result3 := svc.Check(context.Background())
	assert.Equal(t, "healthy", result3.Status)
	assert.Equal(t, 2, callCount)
}

func TestService_LatencyReported(t *testing.T) {
	checkers := map[string]ServiceChecker{
		"postfix": &mockChecker{healthy: true, latencyMs: 12},
		"dovecot": &mockChecker{healthy: true, latencyMs: 3},
		"rspamd":  &mockChecker{healthy: true, latencyMs: 45},
		"redis":   &mockChecker{healthy: true, latencyMs: 7},
	}

	svc := NewService(checkers, "0.1.0")
	result := svc.Check(context.Background())

	serviceMap := serviceSliceToMap(result.Services)
	assert.Equal(t, 12, serviceMap["postfix"].LatencyMs)
	assert.Equal(t, 3, serviceMap["dovecot"].LatencyMs)
	assert.Equal(t, 45, serviceMap["rspamd"].LatencyMs)
	assert.Equal(t, 7, serviceMap["redis"].LatencyMs)
}

func TestService_UptimeIncreases(t *testing.T) {
	checkers := map[string]ServiceChecker{
		"postfix": &mockChecker{healthy: true, latencyMs: 1},
		"dovecot": &mockChecker{healthy: true, latencyMs: 1},
		"rspamd":  &mockChecker{healthy: true, latencyMs: 1},
		"redis":   &mockChecker{healthy: true, latencyMs: 1},
	}

	svc := NewService(checkers, "0.1.0")
	svc.SetStartTime(time.Now().Add(-2 * time.Hour))

	result := svc.Check(context.Background())
	assert.Contains(t, result.Uptime, "h")
}

// serviceSliceToMap converts the services slice to a map for easier assertions.
func serviceSliceToMap(services []ServiceResult) map[string]ServiceResult {
	m := make(map[string]ServiceResult, len(services))
	for _, s := range services {
		m[s.Name] = s
	}
	return m
}
