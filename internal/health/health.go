package health

import (
	"context"
	"database/sql"
	"time"
)

// HealthChecker provides health check functionality
type HealthChecker struct {
	db *sql.DB
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *sql.DB) *HealthChecker {
	return &HealthChecker{db: db}
}

// Check performs a health check
func (hc *HealthChecker) Check(ctx context.Context) error {
	// Check database connectivity
	if err := hc.db.PingContext(ctx); err != nil {
		return err
	}

	// You can add more health checks here
	// - Check if critical tables exist
	// - Check if services are responding
	// - Check memory/CPU usage
	// - Check external dependencies

	return nil
}

// CheckWithTimeout performs a health check with timeout
func (hc *HealthChecker) CheckWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return hc.Check(ctx)
}
