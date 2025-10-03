package services

import (
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// RiskAnalysisService handles risk analysis
type RiskAnalysisService struct {
	profileService *UserProfileService
}

// NewRiskAnalysisService creates a new RiskAnalysisService instance
func NewRiskAnalysisService(
	profileService *UserProfileService,
) *RiskAnalysisService {
	return &RiskAnalysisService{
		profileService: profileService,
	}
}

// CalculateRiskScore calculates risk score for a transaction (Level 4)
func (s *RiskAnalysisService) CalculateRiskScore(hctx hyperion.Context, userID string, amount float64) (riskScore float64, err error) {
	hctx, end := hctx.UseIntercept("RiskAnalysisService", "CalculateRiskScore")
	defer end(&err)

	hctx.Logger().Info("calculating risk score", "user_id", userID, "amount", amount)

	// Get user profile for risk analysis (Level 5)
	profile, err := s.profileService.GetUserProfile(hctx, userID)
	if err != nil {
		hctx.Logger().Error("failed to get user profile", "error", err)
		// Use default high risk score if profile unavailable
		return 80.0, nil
	}

	// Simulate complex risk calculation
	time.Sleep(time.Duration(rand.Intn(25)) * time.Millisecond)

	// Calculate risk score based on user profile and amount
	if profile.IsVerified {
		riskScore = rand.Float64() * 30 // Low risk: 0-30
	} else {
		riskScore = 40 + rand.Float64()*40 // Medium-high risk: 40-80
	}

	// Increase risk for large transactions
	if amount > 1000 {
		riskScore += 10
	}

	hctx.Logger().Info("risk score calculated", "risk_score", riskScore, "user_verified", profile.IsVerified)

	return riskScore, nil
}
