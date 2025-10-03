package hyperion

import (
	"context"
	"testing"
)

// TestNewContextFactory tests the factory constructor.
func TestNewContextFactory(t *testing.T) {
	logger := NewNoOpLogger()
	tracer := NewNoOpTracer()
	db := NewNoOpDatabase()
	meter := NewNoOpMeter()

	factory := NewContextFactory(logger, tracer, db, meter)
	if factory == nil {
		t.Fatal("NewContextFactory should return a factory")
	}

	// Verify factory creates valid contexts
	ctx := factory.New(context.Background())
	if ctx == nil {
		t.Error("Factory.New() should return a context")
	}
	if ctx.Logger() == nil {
		t.Error("Context should have a logger")
	}
	if ctx.Tracer() == nil {
		t.Error("Context should have a tracer")
	}
	if ctx.DB() == nil {
		t.Error("Context should have a database executor")
	}
	if ctx.Meter() == nil {
		t.Error("Context should have a meter")
	}
}
