package services

import (
	"fmt"
	"time"

	"github.com/mapoio/hyperion"
)

// InventoryService handles inventory operations
type InventoryService struct {
	warehouseService *WarehouseService
	// Metrics
	stockCheckCounter    hyperion.Counter   // Total stock checks
	stockCheckDuration   hyperion.Histogram // Stock check duration
	reservationCounter   hyperion.Counter   // Total reservations
	reservationDuration  hyperion.Histogram // Reservation duration
}

// NewInventoryService creates a new InventoryService instance
func NewInventoryService(
	warehouseService *WarehouseService,
	meter hyperion.Meter,
) *InventoryService {
	return &InventoryService{
		warehouseService: warehouseService,
		// Initialize metrics
		stockCheckCounter: meter.Counter("inventory.stock_check.total",
			hyperion.WithMetricDescription("Total number of stock checks"),
			hyperion.WithMetricUnit("1"),
		),
		stockCheckDuration: meter.Histogram("inventory.stock_check.duration",
			hyperion.WithMetricDescription("Stock check duration in milliseconds"),
			hyperion.WithMetricUnit("ms"),
		),
		reservationCounter: meter.Counter("inventory.reservation.total",
			hyperion.WithMetricDescription("Total number of stock reservations"),
			hyperion.WithMetricUnit("1"),
		),
		reservationDuration: meter.Histogram("inventory.reservation.duration",
			hyperion.WithMetricDescription("Stock reservation duration in milliseconds"),
			hyperion.WithMetricUnit("ms"),
		),
	}
}

// CheckStock checks if product is in stock (Level 2)
func (s *InventoryService) CheckStock(hctx hyperion.Context, productID string) (available bool, err error) {
	// Track stock check start time
	startTime := time.Now()

	defer func() {
		// Record processing duration
		duration := float64(time.Since(startTime).Milliseconds())
		status := "success"
		availabilityStatus := "out_of_stock"
		if err != nil {
			status = "error"
		} else if available {
			availabilityStatus = "in_stock"
		}

		s.stockCheckDuration.Record(hctx, duration,
			hyperion.String("service", "inventory"),
			hyperion.String("operation", "check_stock"),
			hyperion.String("status", status),
		)

		s.stockCheckCounter.Add(hctx, 1,
			hyperion.String("service", "inventory"),
			hyperion.String("status", status),
			hyperion.String("availability", availabilityStatus),
		)
	}()

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
	// Track reservation start time
	startTime := time.Now()

	defer func() {
		// Record processing duration
		duration := float64(time.Since(startTime).Milliseconds())
		status := "success"
		if err != nil {
			status = "error"
		}

		s.reservationDuration.Record(hctx, duration,
			hyperion.String("service", "inventory"),
			hyperion.String("operation", "reserve_stock"),
			hyperion.String("status", status),
		)

		s.reservationCounter.Add(hctx, 1,
			hyperion.String("service", "inventory"),
			hyperion.String("status", status),
		)
	}()

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
