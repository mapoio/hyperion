package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// CacheService handles caching operations
type CacheService struct {
	storageService *StorageService
}

// NewCacheService creates a new CacheService instance
func NewCacheService(
	storageService *StorageService,
) *CacheService {
	return &CacheService{
		storageService: storageService,
	}
}

// GetUserProfile retrieves user profile from cache (Level 6)
func (s *CacheService) GetUserProfile(hctx hyperion.Context, userID string) (profile *UserProfile, err error) {
	hctx, end := hctx.UseIntercept("CacheService", "GetUserProfile")
	defer end(&err)

	hctx.Logger().Info("retrieving user profile from cache", "user_id", userID)

	// Check distributed cache storage (Level 7)
	exists, err := s.storageService.KeyExists(hctx, fmt.Sprintf("user_profile:%s", userID))
	if err != nil {
		hctx.Logger().Error("cache storage check failed", "error", err)
		return nil, err
	}

	// Simulate cache lookup
	time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)

	if !exists || rand.Float64() < 0.3 { // 30% cache miss rate
		hctx.Logger().Info("cache miss", "user_id", userID)
		return nil, fmt.Errorf("cache miss")
	}

	// Simulate retrieving cached data
	profile = &UserProfile{
		UserID:     userID,
		IsVerified: rand.Float64() > 0.3,
		TrustScore: rand.Intn(100),
	}

	hctx.Logger().Info("cache hit", "user_id", userID)

	return profile, nil
}

// SetUserProfile stores user profile in cache (Level 6)
func (s *CacheService) SetUserProfile(hctx hyperion.Context, userID string, profile *UserProfile) (err error) {
	hctx, end := hctx.UseIntercept("CacheService", "SetUserProfile")
	defer end(&err)

	hctx.Logger().Info("storing user profile in cache", "user_id", userID)

	// Store in distributed cache storage (Level 7)
	if err := s.storageService.SetKey(hctx, fmt.Sprintf("user_profile:%s", userID), profile); err != nil {
		hctx.Logger().Error("cache storage set failed", "error", err)
		return err
	}

	// Simulate cache set operation
	time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)

	hctx.Logger().Info("user profile stored in cache", "user_id", userID)

	return nil
}
