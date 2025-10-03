package services

import (
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// MappingService handles mapping operations
type MappingService struct {
	monitoringService *MonitoringService
}

// NewMappingService creates a new MappingService instance
func NewMappingService(
	monitoringService *MonitoringService,
) *MappingService {
	return &MappingService{
		monitoringService: monitoringService,
	}
}

// GetMappingData retrieves mapping data (Level 9)
func (s *MappingService) GetMappingData(hctx hyperion.Context, productID string) (hasMappingData bool, err error) {
	hctx, end := hctx.UseIntercept("MappingService", "GetMappingData")
	defer end(&err)

	hctx.Logger().Info("retrieving mapping data", "product_id", productID)

	// Collect metrics (Level 10 - final level!)
	metricsAvailable, err := s.monitoringService.CollectMetrics(hctx)
	if err != nil {
		hctx.Logger().Error("metrics collection failed", "error", err)
		return false, err
	}

	// Simulate mapping data lookup
	time.Sleep(time.Duration(rand.Intn(8)) * time.Millisecond)

	hasMappingData = metricsAvailable && rand.Float64() > 0.05 // 95% have mapping if metrics available

	hctx.Logger().Info("mapping data retrieved", "has_data", hasMappingData, "metrics_available", metricsAvailable)

	return hasMappingData, nil
}
