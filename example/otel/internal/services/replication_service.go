package services

import (
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// ReplicationService handles data replication
type ReplicationService struct {
	healthService   *HealthCheckService
}

// NewReplicationService creates a new ReplicationService instance
func NewReplicationService(
	healthService *HealthCheckService,
) *ReplicationService {
	return &ReplicationService{
		healthService: healthService,
	}
}

// CheckReplicationStatus checks replication status (Level 8)
func (s *ReplicationService) CheckReplicationStatus(hctx hyperion.Context, key string) (isReplicated bool, err error) {
	hctx, end := hctx.UseIntercept("ReplicationService", "CheckReplicationStatus")
	defer end(&err)

	hctx.Logger().Info("checking replication status", "key", key)

	// Check cluster health (Level 9)
	isHealthy, err := s.healthService.CheckClusterHealth(hctx)
	if err != nil {
		hctx.Logger().Error("cluster health check failed", "error", err)
		return false, err
	}

	// Simulate replication status check
	time.Sleep(time.Duration(rand.Intn(8)) * time.Millisecond)

	isReplicated = isHealthy && rand.Float64() > 0.1 // 90% replicated if cluster healthy

	hctx.Logger().Info("replication status checked", "key", key, "replicated", isReplicated, "cluster_healthy", isHealthy)

	return isReplicated, nil
}

// ReplicateData replicates data to other nodes (Level 8)
func (s *ReplicationService) ReplicateData(hctx hyperion.Context, key string, value any) (err error) {
	hctx, end := hctx.UseIntercept("ReplicationService", "ReplicateData")
	defer end(&err)

	hctx.Logger().Info("replicating data", "key", key)

	// Check cluster health (Level 9)
	isHealthy, err := s.healthService.CheckClusterHealth(hctx)
	if err != nil {
		hctx.Logger().Error("cluster health check failed", "error", err)
		return err
	}

	if !isHealthy {
		hctx.Logger().Warn("cluster unhealthy, skipping replication", "key", key)
		return nil
	}

	// Simulate data replication
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

	hctx.Logger().Info("data replicated successfully", "key", key)

	return nil
}
