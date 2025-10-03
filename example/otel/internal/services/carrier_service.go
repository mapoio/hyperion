package services

import (
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// CarrierService handles carrier operations
type CarrierService struct {
	routeService *RouteService
}

// NewCarrierService creates a new CarrierService instance
func NewCarrierService(
	routeService *RouteService,
) *CarrierService {
	return &CarrierService{
		routeService: routeService,
	}
}

// CheckCapacity checks carrier capacity (Level 5)
func (s *CarrierService) CheckCapacity(hctx hyperion.Context, productID string) (hasCapacity bool, err error) {
	hctx, end := hctx.UseIntercept("CarrierService", "CheckCapacity")
	defer end(&err)

	hctx.Logger().Info("checking carrier capacity", "product_id", productID)

	// Check optimal route (Level 6)
	hasRoute, err := s.routeService.FindOptimalRoute(hctx, productID)
	if err != nil {
		hctx.Logger().Error("route finding failed", "error", err)
		return false, err
	}

	// Simulate capacity check
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

	hasCapacity = hasRoute && rand.Float64() > 0.15 // 85% have capacity if route exists

	hctx.Logger().Info("carrier capacity checked", "has_capacity", hasCapacity, "has_route", hasRoute)

	return hasCapacity, nil
}
