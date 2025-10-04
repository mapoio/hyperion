package services

import (
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// RouteService handles route operations
type RouteService struct {
	geoService     *GeoService
}

// NewRouteService creates a new RouteService instance
func NewRouteService(
	geoService *GeoService,
) *RouteService {
	return &RouteService{
		geoService: geoService,
	}
}

// FindOptimalRoute finds optimal shipping route (Level 6)
func (s *RouteService) FindOptimalRoute(hctx hyperion.Context, productID string) (hasRoute bool, err error) {
	hctx, end := hctx.UseIntercept("RouteService", "FindOptimalRoute")
	defer end(&err)

	hctx.Logger().Info("finding optimal route", "product_id", productID)

	// Get geo location data (Level 7)
	hasLocation, err := s.geoService.GetLocationData(hctx, productID)
	if err != nil {
		hctx.Logger().Error("geo location lookup failed", "error", err)
		return false, err
	}

	// Simulate route calculation
	time.Sleep(time.Duration(rand.Intn(15)) * time.Millisecond)

	hasRoute = hasLocation && rand.Float64() > 0.1 // 90% have route if location exists

	hctx.Logger().Info("optimal route search completed", "found", hasRoute, "has_location", hasLocation)

	return hasRoute, nil
}
