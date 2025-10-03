package services

import (
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// ShippingService handles shipping operations
type ShippingService struct {
	carrierService *CarrierService
}

// NewShippingService creates a new ShippingService instance
func NewShippingService(
	carrierService *CarrierService,
) *ShippingService {
	return &ShippingService{
		carrierService: carrierService,
	}
}

// CheckShippingAvailability checks if shipping is available (Level 4)
func (s *ShippingService) CheckShippingAvailability(hctx hyperion.Context, productID string) (available bool, err error) {
	hctx, end := hctx.UseIntercept("ShippingService", "CheckShippingAvailability")
	defer end(&err)

	hctx.Logger().Info("checking shipping availability", "product_id", productID)

	// Check carrier capacity (Level 5)
	hasCapacity, err := s.carrierService.CheckCapacity(hctx, productID)
	if err != nil {
		hctx.Logger().Error("carrier capacity check failed", "error", err)
		return false, err
	}

	// Simulate shipping zone check
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

	available = hasCapacity && rand.Float64() > 0.1 // 90% availability

	hctx.Logger().Info("shipping availability checked", "available", available, "carrier_capacity", hasCapacity)

	return available, nil
}
