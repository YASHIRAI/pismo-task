package health

import (
	"context"
	"database/sql"
	"time"
)

// HealthChecker provides health check functionality for the application.
// It can be extended to check database connectivity, service dependencies, and system resources.
type HealthChecker struct {
	db *sql.DB
}

// NewHealthChecker creates a new health checker instance.
// It takes a database connection and returns a configured HealthChecker.
func NewHealthChecker(db *sql.DB) *HealthChecker {
	return &HealthChecker{db: db}
}

// Check performs a comprehensive health check.
// It verifies database connectivity and can be extended to check other system components.
// Returns an error if any health check fails.
func (hc *HealthChecker) Check(ctx context.Context) error {
	if err := hc.db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

// CheckWithTimeout performs a health check with a specified timeout duration.
// It creates a context with timeout and delegates to the main Check method.
// Returns an error if the health check fails or times out.
func (hc *HealthChecker) CheckWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return hc.Check(ctx)
}
