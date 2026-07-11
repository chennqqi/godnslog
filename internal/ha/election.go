package ha

import (
	"context"
	"time"
)

// ElectLeader attempts to acquire the leader lock for the given node.
func (s *Service) ElectLeader(ctx context.Context, nodeID string) (bool, error) {
	return s.store.TryAcquireLock(ctx, nodeID, 30*time.Second)
}

// IsLeader checks if the given node is currently the leader.
func (s *Service) IsLeader(ctx context.Context, nodeID string) (bool, error) {
	leader, err := s.store.GetLeader(ctx)
	if err != nil {
		return false, err
	}
	if leader == nil {
		return false, nil
	}
	return leader.LeaderID == nodeID && leader.LeaseEnd.After(time.Now()), nil
}

// GetLeaderID returns the current leader's node ID, or empty string if none.
func (s *Service) GetLeaderID(ctx context.Context) (string, error) {
	leader, err := s.store.GetLeader(ctx)
	if err != nil {
		return "", err
	}
	if leader == nil || leader.LeaseEnd.Before(time.Now()) {
		return "", nil
	}
	return leader.LeaderID, nil
}

// Resign releases the leader lock held by the given node.
func (s *Service) Resign(ctx context.Context, nodeID string) error {
	return s.store.ReleaseLock(ctx, nodeID)
}
