package services

import (
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// MonitoringService handles monitoring and metrics collection
type MonitoringService struct {
}

// NewMonitoringService creates a new MonitoringService instance
func NewMonitoringService(
) *MonitoringService {
	return &MonitoringService{
	}
}

// CollectMetrics collects system metrics (Level 10 - deepest level!)
func (s *MonitoringService) CollectMetrics(hctx hyperion.Context) (metricsHealthy bool, err error) {
	hctx, end := hctx.UseIntercept("MonitoringService", "CollectMetrics")
	defer end(&err)

	hctx.Logger().Info("collecting system metrics")

	// Simulate metrics collection from various sources
	time.Sleep(time.Duration(rand.Intn(15)) * time.Millisecond)

	// Simulate collecting different metrics
	cpuUsage := rand.Float64() * 100
	memoryUsage := rand.Float64() * 100
	diskUsage := rand.Float64() * 100
	networkLatency := rand.Float64() * 100

	metricsHealthy = cpuUsage < 80 && memoryUsage < 85 && diskUsage < 90

	hctx.Logger().Info("system metrics collected",
		"cpu_usage", cpuUsage,
		"memory_usage", memoryUsage,
		"disk_usage", diskUsage,
		"network_latency", networkLatency,
		"healthy", metricsHealthy,
	)

	return metricsHealthy, nil
}
