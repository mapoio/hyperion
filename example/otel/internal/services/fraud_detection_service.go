package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// FraudDetectionService handles fraud detection
type FraudDetectionService struct {
	riskService *RiskAnalysisService
	// Metrics
	fraudCheckCounter  hyperion.Counter   // Total fraud checks performed
	fraudCheckDuration hyperion.Histogram // Fraud check duration
	fraudDetectedGauge hyperion.UpDownCounter // Current fraud rate gauge
	riskScoreHistogram hyperion.Histogram // Risk score distribution
}

// NewFraudDetectionService creates a new FraudDetectionService instance
func NewFraudDetectionService(
	riskService *RiskAnalysisService,
	meter hyperion.Meter,
) *FraudDetectionService {
	return &FraudDetectionService{
		riskService: riskService,
		// Initialize metrics
		fraudCheckCounter: meter.Counter("fraud.check.total",
			hyperion.WithMetricDescription("Total number of fraud checks performed"),
			hyperion.WithMetricUnit("1"),
		),
		fraudCheckDuration: meter.Histogram("fraud.check.duration",
			hyperion.WithMetricDescription("Fraud check duration in milliseconds"),
			hyperion.WithMetricUnit("ms"),
		),
		fraudDetectedGauge: meter.UpDownCounter("fraud.detected.rate",
			hyperion.WithMetricDescription("Fraud detection rate"),
			hyperion.WithMetricUnit("1"),
		),
		riskScoreHistogram: meter.Histogram("fraud.risk_score",
			hyperion.WithMetricDescription("Risk score distribution"),
			hyperion.WithMetricUnit("score"),
		),
	}
}

// CheckTransaction checks if a transaction is fraudulent (Level 3)
func (s *FraudDetectionService) CheckTransaction(hctx hyperion.Context, userID string, amount float64) (isSafe bool, err error) {
	// Track fraud check start time
	startTime := time.Now()

	defer func() {
		// Record processing duration
		duration := float64(time.Since(startTime).Milliseconds())
		status := "success"
		if err != nil {
			status = "error"
		}

		s.fraudCheckDuration.Record(hctx, duration,
			hyperion.String("service", "fraud_detection"),
			hyperion.String("operation", "check_transaction"),
			hyperion.String("status", status),
		)

		// Record fraud check counter
		fraudStatus := "safe"
		if !isSafe {
			fraudStatus = "fraudulent"
			// Increment fraud detected gauge
			s.fraudDetectedGauge.Add(hctx, int64(1),
				hyperion.String("service", "fraud_detection"),
			)
		}

		s.fraudCheckCounter.Add(hctx, 1,
			hyperion.String("service", "fraud_detection"),
			hyperion.String("status", status),
			hyperion.String("fraud_status", fraudStatus),
		)
	}()

	hctx, end := hctx.UseIntercept("FraudDetectionService", "CheckTransaction")
	defer end(&err)

	hctx.Logger().Info("checking transaction for fraud", "user_id", userID, "amount", amount)

	// Analyze risk score (Level 4)
	riskScore, err := s.riskService.CalculateRiskScore(hctx, userID, amount)
	if err != nil {
		hctx.Logger().Error("risk analysis failed", "error", err)
		return false, fmt.Errorf("risk analysis failed: %w", err)
	}

	// Record risk score
	s.riskScoreHistogram.Record(hctx, float64(riskScore),
		hyperion.String("service", "fraud_detection"),
	)

	// Simulate fraud detection model
	time.Sleep(time.Duration(rand.Intn(15)) * time.Millisecond)

	isSafe = riskScore < 70

	hctx.Logger().Info("fraud check completed", "risk_score", riskScore, "is_safe", isSafe)

	return isSafe, nil
}
