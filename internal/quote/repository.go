package quote

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	FindByUserID(ctx context.Context, userID string, limit int) ([]Quote, error)
	Create(ctx context.Context, q *Quote) error
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) FindByUserID(ctx context.Context, userID string, limit int) ([]Quote, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT id, user_id, pickup, destination, price, created_at
		FROM quotes
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotes []Quote
	for rows.Next() {
		var q Quote
		var createdAt time.Time
		if err := rows.Scan(
			&q.ID,
			&q.UserID,
			&q.Pickup,
			&q.Destination,
			&q.Price,
			&createdAt,
		); err != nil {
			return nil, err
		}
		q.CreatedAt = createdAt.Format(time.RFC3339)
		quotes = append(quotes, q)
	}

	return quotes, nil
}

func (r *PostgresRepository) Create(ctx context.Context, q *Quote) error {
	query := `
		INSERT INTO quotes (user_id, pickup, destination, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	var createdAt time.Time
	err := r.pool.QueryRow(
		ctx,
		query,
		q.UserID,
		q.Pickup,
		q.Destination,
		q.Price,
	).Scan(&q.ID, &createdAt)

	if err != nil {
		return fmt.Errorf("erro ao salvar orçamento: %w", err)
	}

	q.CreatedAt = createdAt.Format(time.RFC3339)
	return nil
}

// In-memory fallback
type MemoryRepository struct {
	mu     sync.RWMutex
	quotes map[string]*Quote
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		quotes: make(map[string]*Quote),
	}
}

func (r *MemoryRepository) FindByUserID(ctx context.Context, userID string, limit int) ([]Quote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []Quote
	for _, q := range r.quotes {
		if q.UserID == userID {
			list = append(list, *q)
		}
	}
	return list, nil
}

func (r *MemoryRepository) Create(ctx context.Context, q *Quote) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	q.ID = fmt.Sprintf("quote_%d", time.Now().UnixNano())
	q.CreatedAt = time.Now().Format(time.RFC3339)
	r.quotes[q.ID] = q
	return nil
}
