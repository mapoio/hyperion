package services

import (
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// UserProfile represents a user profile
type UserProfile struct {
	UserID     string
	IsVerified bool
	TrustScore int
}

// UserProfileService handles user profile operations
type UserProfileService struct {
	cacheService *CacheService
}

// NewUserProfileService creates a new UserProfileService instance
func NewUserProfileService(
	cacheService *CacheService,
) *UserProfileService {
	return &UserProfileService{
		cacheService: cacheService,
	}
}

// GetUserProfile retrieves user profile (Level 5)
func (s *UserProfileService) GetUserProfile(hctx hyperion.Context, userID string) (profile *UserProfile, err error) {
	hctx, end := hctx.UseIntercept("UserProfileService", "GetUserProfile")
	defer end(&err)

	hctx.Logger().Info("retrieving user profile", "user_id", userID)

	// Try to get from cache first (Level 6)
	cachedProfile, err := s.cacheService.GetUserProfile(hctx, userID)
	if err == nil && cachedProfile != nil {
		hctx.Logger().Info("user profile found in cache", "user_id", userID)
		return cachedProfile, nil
	}

	// Cache miss - simulate database query
	time.Sleep(time.Duration(rand.Intn(30)) * time.Millisecond)

	profile = &UserProfile{
		UserID:     userID,
		IsVerified: rand.Float64() > 0.3, // 70% verified users
		TrustScore: rand.Intn(100),
	}

	// Store in cache (Level 6)
	if err := s.cacheService.SetUserProfile(hctx, userID, profile); err != nil {
		hctx.Logger().Warn("failed to cache user profile", "error", err)
		// Don't fail the request if caching fails
	}

	hctx.Logger().Info("user profile retrieved", "user_id", userID, "verified", profile.IsVerified, "trust_score", profile.TrustScore)

	return profile, nil
}
