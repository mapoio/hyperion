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
	// Metrics
	stockLevelGauge       hyperion.UpDownCounter // Current stock level gauge
	stockQueryCounter     hyperion.Counter   // Total stock queries
	stockQueryDuration    hyperion.Histogram // Stock query duration
	reservationCounter    hyperion.Counter   // Total reservations
	reservationDuration   hyperion.Histogram // Reservation duration
	reservationFailures   hyperion.Counter   // Reservation failures
}

// NewWarehouseService creates a new WarehouseService instance
func NewWarehouseService(
	shippingService *ShippingService,
	meter hyperion.Meter,
) *WarehouseService {
	return &WarehouseService{
		shippingService: shippingService,
		// Initialize metrics
		stockLevelGauge: meter.UpDownCounter("warehouse.stock_level",
			hyperion.WithMetricDescription("Current stock level"),
			hyperion.WithMetricUnit("units"),
		),
		stockQueryCounter: meter.Counter("warehouse.stock_query.total",
			hyperion.WithMetricDescription("Total number of stock queries"),
			hyperion.WithMetricUnit("1"),
		),
		stockQueryDuration: meter.Histogram("warehouse.stock_query.duration",
			hyperion.WithMetricDescription("Stock query duration in milliseconds"),
			hyperion.WithMetricUnit("ms"),
		),
		reservationCounter: meter.Counter("warehouse.reservation.total",
			hyperion.WithMetricDescription("Total number of reservations"),
			hyperion.WithMetricUnit("1"),
		),
		reservationDuration: meter.Histogram("warehouse.reservation.duration",
			hyperion.WithMetricDescription("Reservation duration in milliseconds"),
			hyperion.WithMetricUnit("ms"),
		),
		reservationFailures: meter.Counter("warehouse.reservation.failures",
			hyperion.WithMetricDescription("Number of reservation failures"),
			hyperion.WithMetricUnit("1"),
		),
	}
}

// GetStockLevel retrieves stock level for a product (Level 3)
func (s *WarehouseService) GetStockLevel(hctx hyperion.Context, productID string) (stockLevel int, err error) {
	// Track stock query start time
	startTime := time.Now()

	defer func() {
		// Record processing duration
		duration := float64(time.Since(startTime).Milliseconds())
		status := "success"
		if err != nil {
			status = "error"
		}

		s.stockQueryDuration.Record(hctx, duration,
			hyperion.String("service", "warehouse"),
			hyperion.String("operation", "get_stock_level"),
			hyperion.String("status", status),
		)

		s.stockQueryCounter.Add(hctx, 1,
			hyperion.String("service", "warehouse"),
			hyperion.String("status", status),
		)

		// Update stock level gauge (if successful)
		if err == nil {
			s.stockLevelGauge.Add(hctx, int64(stockLevel),
				hyperion.String("service", "warehouse"),
				hyperion.String("product_id", productID),
			)
		}
	}()

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
	// Track reservation start time
	startTime := time.Now()

	defer func() {
		// Record processing duration
		duration := float64(time.Since(startTime).Milliseconds())
		status := "success"
		if err != nil {
			status = "error"
			// Record reservation failure
			s.reservationFailures.Add(hctx, 1,
				hyperion.String("service", "warehouse"),
				hyperion.String("product_id", productID),
			)
		}

		s.reservationDuration.Record(hctx, duration,
			hyperion.String("service", "warehouse"),
			hyperion.String("operation", "reserve_stock"),
			hyperion.String("status", status),
		)

		s.reservationCounter.Add(hctx, 1,
			hyperion.String("service", "warehouse"),
			hyperion.String("status", status),
		)
	}()

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
