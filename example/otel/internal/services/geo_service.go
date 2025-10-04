package services

import (
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// GeoService handles geolocation operations
type GeoService struct {
	coordinateService *CoordinateService
}

// NewGeoService creates a new GeoService instance
func NewGeoService(
	coordinateService *CoordinateService,
) *GeoService {
	return &GeoService{
		coordinateService: coordinateService,
	}
}

// GetLocationData retrieves location data (Level 7)
func (s *GeoService) GetLocationData(hctx hyperion.Context, productID string) (hasLocation bool, err error) {
	hctx, end := hctx.UseIntercept("GeoService", "GetLocationData")
	defer end(&err)

	hctx.Logger().Info("retrieving location data", "product_id", productID)

	// Get precise coordinates (Level 8)
	hasCoordinates, err := s.coordinateService.GetCoordinates(hctx, productID)
	if err != nil {
		hctx.Logger().Error("coordinate retrieval failed", "error", err)
		return false, err
	}

	// Simulate geo lookup
	time.Sleep(time.Duration(rand.Intn(12)) * time.Millisecond)

	hasLocation = hasCoordinates && rand.Float64() > 0.05 // 95% have location if coordinates exist

	hctx.Logger().Info("location data retrieved", "has_location", hasLocation, "has_coordinates", hasCoordinates)

	return hasLocation, nil
}
