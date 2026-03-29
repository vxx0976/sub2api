package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

// Sub2apipayService provides integration with sub2apipay payment system
// It reuses the main database connection but switches to the sub2apipay database
type Sub2apipayService struct {
	db *sql.DB
}

// NewSub2apipayService creates a new Sub2apipayService
// It derives the sub2apipay DSN from the main config
func NewSub2apipayService(cfg *config.Config) (*Sub2apipayService, error) {
	if !cfg.Sub2apipay.Enabled {
		return nil, nil
	}

	dbName := cfg.Sub2apipay.DatabaseName
	if dbName == "" {
		dbName = "sub2apipay"
	}

	// Build DSN from main database config, but change the database name
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		dbName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sub2apipay database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sub2apipay database: %w", err)
	}

	return &Sub2apipayService{db: db}, nil
}

// Close closes the database connection
func (s *Sub2apipayService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// SumRechargeByUserIDs returns total recharge amount for given user IDs
// This queries the sub2apipay orders table directly
func (s *Sub2apipayService) SumRechargeByUserIDs(ctx context.Context, userIDs []int64) (float64, error) {
	if len(userIDs) == 0 {
		return 0, nil
	}

	// Build placeholders for IN clause
	placeholders := make([]byte, 0, len(userIDs)*4)
	args := make([]interface{}, 0, len(userIDs))
	for i, id := range userIDs {
		if i > 0 {
			placeholders = append(placeholders, ',', ' ')
		}
		placeholders = append(placeholders, '$')
		placeholders = append(placeholders, []byte(fmt.Sprintf("%d", i+1))...)
		args = append(args, id)
	}

	// Query paid/completed orders only
	// Status: PAID, RECHARGING, COMPLETED, REFUNDING, REFUNDED, REFUND_FAILED
	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(amount), 0)
		FROM orders
		WHERE user_id IN (%s)
		AND status IN ('PAID', 'RECHARGING', 'COMPLETED', 'REFUNDING', 'REFUNDED', 'REFUND_FAILED')
	`, string(placeholders))

	var total sql.NullFloat64
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to query recharge sum: %w", err)
	}

	if !total.Valid {
		return 0, nil
	}

	return total.Float64, nil
}
