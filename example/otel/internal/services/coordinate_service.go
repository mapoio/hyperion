package services

import (
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// CoordinateService handles coordinate operations
type CoordinateService struct {
	mappingService *MappingService
}

// NewCoordinateService creates a new CoordinateService instance
func NewCoordinateService(
	mappingService *MappingService,
) *CoordinateService {
	return &CoordinateService{
		mappingService: mappingService,
	}
}

// GetCoordinates retrieves coordinates (Level 8)
func (s *CoordinateService) GetCoordinates(hctx hyperion.Context, productID string) (hasCoordinates bool, err error) {
	hctx, end := hctx.UseIntercept("CoordinateService", "GetCoordinates")
	defer end(&err)

	hctx.Logger().Info("retrieving coordinates", "product_id", productID)

	// Get mapping data (Level 9)
	hasMappingData, err := s.mappingService.GetMappingData(hctx, productID)
	if err != nil {
		hctx.Logger().Error("mapping data retrieval failed", "error", err)
		return false, err
	}

	// Simulate coordinate lookup
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

	hasCoordinates = hasMappingData && rand.Float64() > 0.05 // 95% have coordinates if mapping exists

	// Simulate latitude/longitude
	latitude := rand.Float64()*180 - 90
	longitude := rand.Float64()*360 - 180

	hctx.Logger().Info("coordinates retrieved", "has_coordinates", hasCoordinates, "lat", latitude, "lon", longitude)

	return hasCoordinates, nil
}
