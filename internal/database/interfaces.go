package database

import "context"

// HealthChecker is the minimal database contract required by health checks.
type HealthChecker interface {
	Ping(ctx context.Context) error
}
