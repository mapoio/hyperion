package services

import (
	"go.uber.org/fx"
)

// Module provides all services with proper dependency injection
var Module = fx.Module("services",
	// Level 10 (deepest)
	fx.Provide(NewMonitoringService),

	// Level 9
	fx.Provide(NewHealthCheckService),
	fx.Provide(NewMappingService),

	// Level 8
	fx.Provide(NewCoordinateService),
	fx.Provide(NewReplicationService),

	// Level 7
	fx.Provide(NewGeoService),
	fx.Provide(NewStorageService),

	// Level 6
	fx.Provide(NewCacheService),
	fx.Provide(NewRouteService),

	// Level 5
	fx.Provide(NewCarrierService),
	fx.Provide(NewUserProfileService),
	fx.Provide(NewTemplateService),

	// Level 4
	fx.Provide(NewEmailService),
	fx.Provide(NewSMSService),
	fx.Provide(NewRiskAnalysisService),
	fx.Provide(NewShippingService),

	// Level 3
	fx.Provide(NewFraudDetectionService),
	fx.Provide(NewNotificationService),
	fx.Provide(NewWarehouseService),

	// Level 2
	fx.Provide(NewPaymentService),
	fx.Provide(NewInventoryService),

	// Level 1 (entry point)
	fx.Provide(NewOrderService),
)
