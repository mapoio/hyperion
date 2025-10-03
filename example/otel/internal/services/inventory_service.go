package services

import (
	"fmt"

	"github.com/mapoio/hyperion"
)

// InventoryService handles inventory operations
type InventoryService struct {
	warehouseService *WarehouseService
}

// NewInventoryService creates a new InventoryService instance
func NewInventoryService(
	warehouseService *WarehouseService,
) *InventoryService {
	return &InventoryService{
		warehouseService: warehouseService,
	}
}

// CheckStock checks if product is in stock (Level 2)
func (s *InventoryService) CheckStock(hctx hyperion.Context, productID string) (available bool, err error) {
	hctx, end := hctx.UseIntercept("InventoryService", "CheckStock")
	defer end(&err)

	hctx.Logger().Info("checking stock", "product_id", productID)

	// Query warehouse for stock levels (Level 3)
	stockLevel, err := s.warehouseService.GetStockLevel(hctx, productID)
	if err != nil {
		hctx.Logger().Error("warehouse query failed", "error", err)
		return false, fmt.Errorf("warehouse query failed: %w", err)
	}

	available = stockLevel > 0

	hctx.Logger().Info("stock check completed", "product_id", productID, "stock_level", stockLevel, "available", available)

	return available, nil
}

// ReserveStock reserves stock for an order (Level 2)
func (s *InventoryService) ReserveStock(hctx hyperion.Context, productID string, quantity int) (err error) {
	hctx, end := hctx.UseIntercept("InventoryService", "ReserveStock")
	defer end(&err)

	hctx.Logger().Info("reserving stock", "product_id", productID, "quantity", quantity)

	// Reserve stock in warehouse (Level 3)
	if err := s.warehouseService.ReserveStock(hctx, productID, quantity); err != nil {
		hctx.Logger().Error("warehouse reservation failed", "error", err)
		return fmt.Errorf("warehouse reservation failed: %w", err)
	}

	hctx.Logger().Info("stock reserved successfully", "product_id", productID, "quantity", quantity)

	return nil
}
