package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// WarehouseService handles warehouse operations
type WarehouseService struct {
	shippingService *ShippingService
}

// NewWarehouseService creates a new WarehouseService instance
func NewWarehouseService(
	shippingService *ShippingService,
) *WarehouseService {
	return &WarehouseService{
		shippingService: shippingService,
	}
}

// GetStockLevel retrieves stock level for a product (Level 3)
func (s *WarehouseService) GetStockLevel(hctx hyperion.Context, productID string) (stockLevel int, err error) {
	hctx, end := hctx.UseIntercept("WarehouseService", "GetStockLevel")
	defer end(&err)

	hctx.Logger().Info("querying stock level", "product_id", productID)

	// Check shipping availability (Level 4)
	canShip, err := s.shippingService.CheckShippingAvailability(hctx, productID)
	if err != nil {
		hctx.Logger().Warn("shipping availability check failed", "error", err)
		// Continue anyway, not critical for stock level
	}

	// Simulate database query
	time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)

	// Random stock level between 0-100
	stockLevel = rand.Intn(101)

	// If shipping not available, reduce perceived stock
	if !canShip {
		stockLevel = stockLevel / 2
		hctx.Logger().Info("reduced stock level due to shipping constraints", "original", stockLevel*2, "adjusted", stockLevel)
	}

	hctx.Logger().Info("stock level retrieved", "product_id", productID, "stock_level", stockLevel)

	return stockLevel, nil
}

// ReserveStock reserves stock in warehouse (Level 3)
func (s *WarehouseService) ReserveStock(hctx hyperion.Context, productID string, quantity int) (err error) {
	hctx, end := hctx.UseIntercept("WarehouseService", "ReserveStock")
	defer end(&err)

	hctx.Logger().Info("reserving warehouse stock", "product_id", productID, "quantity", quantity)

	// Simulate warehouse reservation
	time.Sleep(time.Duration(rand.Intn(15)) * time.Millisecond)

	// Randomly fail 5% of the time
	if rand.Float64() < 0.05 {
		err := fmt.Errorf("warehouse reservation failed: system error")
		hctx.Logger().Error("warehouse reservation failed", "error", err)
		return err
	}

	hctx.Logger().Info("warehouse stock reserved successfully", "product_id", productID, "quantity", quantity)

	return nil
}
