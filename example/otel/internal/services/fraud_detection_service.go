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
}

// NewFraudDetectionService creates a new FraudDetectionService instance
func NewFraudDetectionService(
	riskService *RiskAnalysisService,
) *FraudDetectionService {
	return &FraudDetectionService{
		riskService: riskService,
	}
}

// CheckTransaction checks if a transaction is fraudulent (Level 3)
func (s *FraudDetectionService) CheckTransaction(hctx hyperion.Context, userID string, amount float64) (isSafe bool, err error) {
	hctx, end := hctx.UseIntercept("FraudDetectionService", "CheckTransaction")
	defer end(&err)

	hctx.Logger().Info("checking transaction for fraud", "user_id", userID, "amount", amount)

	// Analyze risk score (Level 4)
	riskScore, err := s.riskService.CalculateRiskScore(hctx, userID, amount)
	if err != nil {
		hctx.Logger().Error("risk analysis failed", "error", err)
		return false, fmt.Errorf("risk analysis failed: %w", err)
	}

	// Simulate fraud detection model
	time.Sleep(time.Duration(rand.Intn(15)) * time.Millisecond)

	isSafe = riskScore < 70

	hctx.Logger().Info("fraud check completed", "risk_score", riskScore, "is_safe", isSafe)

	return isSafe, nil
}
