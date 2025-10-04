package hyperion

import (
	"go.uber.org/fx"
)

// ProviderParams represents the input parameters for DefaultProviders
type ProviderParams struct {
	fx.In

	Config     Config     `optional:"true"`
	Logger     Logger     `optional:"true"`
	Tracer     Tracer     `optional:"true"`
	Meter      Meter      `optional:"true"`
	Database   Database   `optional:"true"`
	Cache      Cache      `optional:"true"`
	UnitOfWork UnitOfWork `optional:"true"`
}

// ProviderResults represents the output results from DefaultProviders
type ProviderResults struct {
	fx.Out

	Config     Config
	Logger     Logger
	Tracer     Tracer
	Meter      Meter
	Database   Database
	Cache      Cache
	UnitOfWork UnitOfWork
}

// DefaultProviders provides all default NoOp implementations with optional overrides
func DefaultProviders(params ProviderParams) ProviderResults {
	// Use provided implementations or fall back to NoOp
	config := params.Config
	if config == nil {
		config = NewNoOpConfig()
	}

	logger := params.Logger
	if logger == nil {
		logger = NewNoOpLogger()
	}

	tracer := params.Tracer
	if tracer == nil {
		tracer = NewNoOpTracer()
	}

	meter := params.Meter
	if meter == nil {
		meter = NewNoOpMeter()
	}

	database := params.Database
	if database == nil {
		database = NewNoOpDatabase()
	}

	cache := params.Cache
	if cache == nil {
		cache = NewNoOpCache()
	}

	unitOfWork := params.UnitOfWork
	if unitOfWork == nil {
		unitOfWork = NewNoOpUnitOfWork()
	}

	return ProviderResults{
		Config:     config,
		Logger:     logger,
		Tracer:     tracer,
		Meter:      meter,
		Database:   database,
		Cache:      cache,
		UnitOfWork: unitOfWork,
	}
}
