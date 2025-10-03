package services

import (
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// HealthCheckService handles health checks
type HealthCheckService struct {
	monitoringService *MonitoringService
}

// NewHealthCheckService creates a new HealthCheckService instance
func NewHealthCheckService(
	monitoringService *MonitoringService,
) *HealthCheckService {
	return &HealthCheckService{
		monitoringService: monitoringService,
	}
}

// CheckClusterHealth checks cluster health (Level 9)
func (s *HealthCheckService) CheckClusterHealth(hctx hyperion.Context) (isHealthy bool, err error) {
	hctx, end := hctx.UseIntercept("HealthCheckService", "CheckClusterHealth")
	defer end(&err)

	hctx.Logger().Info("checking cluster health")

	// Get monitoring metrics (Level 10 - final level!)
	metricsHealthy, err := s.monitoringService.CollectMetrics(hctx)
	if err != nil {
		hctx.Logger().Error("metrics collection failed", "error", err)
		return false, err
	}

	// Simulate health check
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

	isHealthy = metricsHealthy && rand.Float64() > 0.05 // 95% healthy if metrics ok

	hctx.Logger().Info("cluster health checked", "is_healthy", isHealthy, "metrics_healthy", metricsHealthy)

	return isHealthy, nil
}
