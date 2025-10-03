package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// StorageService handles distributed storage operations
type StorageService struct {
	replicationService *ReplicationService
}

// NewStorageService creates a new StorageService instance
func NewStorageService(
	replicationService *ReplicationService,
) *StorageService {
	return &StorageService{
		replicationService: replicationService,
	}
}

// KeyExists checks if a key exists in storage (Level 7)
func (s *StorageService) KeyExists(hctx hyperion.Context, key string) (exists bool, err error) {
	hctx, end := hctx.UseIntercept("StorageService", "KeyExists")
	defer end(&err)

	hctx.Logger().Info("checking key existence", "key", key)

	// Check replication status (Level 8)
	isReplicated, err := s.replicationService.CheckReplicationStatus(hctx, key)
	if err != nil {
		hctx.Logger().Warn("replication status check failed", "error", err)
		// Continue anyway
	}

	// Simulate storage lookup
	time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)

	exists = isReplicated && rand.Float64() > 0.2 // 80% exists if replicated

	hctx.Logger().Info("key existence checked", "key", key, "exists", exists, "replicated", isReplicated)

	return exists, nil
}

// SetKey sets a key in storage (Level 7)
func (s *StorageService) SetKey(hctx hyperion.Context, key string, value interface{}) (err error) {
	hctx, end := hctx.UseIntercept("StorageService", "SetKey")
	defer end(&err)

	hctx.Logger().Info("setting key in storage", "key", key)

	// Simulate storage write
	time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)

	// Replicate to other nodes (Level 8)
	if err := s.replicationService.ReplicateData(hctx, key, value); err != nil {
		hctx.Logger().Warn("data replication failed", "error", err)
		// Don't fail the write if replication fails
	}

	hctx.Logger().Info("key set in storage", "key", key)

	return nil
}

// LoadTemplate loads a template from storage (Level 6)
func (s *StorageService) LoadTemplate(hctx hyperion.Context, templateName string) (template string, err error) {
	hctx, end := hctx.UseIntercept("StorageService", "LoadTemplate")
	defer end(&err)

	hctx.Logger().Info("loading template", "template_name", templateName)

	// Check replication (Level 8)
	key := fmt.Sprintf("template:%s", templateName)
	isReplicated, err := s.replicationService.CheckReplicationStatus(hctx, key)
	if err != nil {
		hctx.Logger().Warn("template replication check failed", "error", err)
	}

	// Simulate template loading
	time.Sleep(time.Duration(rand.Intn(8)) * time.Millisecond)

	template = fmt.Sprintf("[%s Template]", templateName)

	hctx.Logger().Info("template loaded", "template_name", templateName, "replicated", isReplicated)

	return template, nil
}
