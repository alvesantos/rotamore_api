package vehicle

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	FindByUserID(ctx context.Context, userID string) ([]Vehicle, error)
	Create(ctx context.Context, v *Vehicle) error
	Delete(ctx context.Context, id, userID string) error
	SetActive(ctx context.Context, id, userID string) error
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) FindByUserID(ctx context.Context, userID string) ([]Vehicle, error) {
	query := `
		SELECT id, user_id, name, brand, plate, color, year, is_active, created_at, updated_at
		FROM vehicles
		WHERE user_id = $1
		ORDER BY is_active DESC, created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []Vehicle
	for rows.Next() {
		var v Vehicle
		var createdAt, updatedAt time.Time
		if err := rows.Scan(
			&v.ID,
			&v.UserID,
			&v.Name,
			&v.Brand,
			&v.Plate,
			&v.Color,
			&v.Year,
			&v.IsActive,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		v.CreatedAt = createdAt.Format(time.RFC3339)
		v.UpdatedAt = updatedAt.Format(time.RFC3339)
		vehicles = append(vehicles, v)
	}

	return vehicles, nil
}

func (r *PostgresRepository) Create(ctx context.Context, v *Vehicle) error {
	v.Plate = strings.ToUpper(strings.TrimSpace(v.Plate))

	// Check if this is the first vehicle for user, if so make it active
	var count int
	_ = r.pool.QueryRow(ctx, "SELECT count(*) FROM vehicles WHERE user_id = $1", v.UserID).Scan(&count)
	if count == 0 {
		v.IsActive = true
	}

	query := `
		INSERT INTO vehicles (user_id, name, brand, plate, color, year, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err := r.pool.QueryRow(
		ctx,
		query,
		v.UserID,
		v.Name,
		v.Brand,
		v.Plate,
		v.Color,
		v.Year,
		v.IsActive,
	).Scan(&v.ID, &createdAt, &updatedAt)

	if err != nil {
		return fmt.Errorf("erro ao cadastrar veículo: %w", err)
	}

	v.CreatedAt = createdAt.Format(time.RFC3339)
	v.UpdatedAt = updatedAt.Format(time.RFC3339)
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id, userID string) error {
	query := `DELETE FROM vehicles WHERE id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, id, userID)
	return err
}

func (r *PostgresRepository) SetActive(ctx context.Context, id, userID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Set all inactive
	if _, err := tx.Exec(ctx, "UPDATE vehicles SET is_active = false WHERE user_id = $1", userID); err != nil {
		return err
	}

	// Set target active
	if _, err := tx.Exec(ctx, "UPDATE vehicles SET is_active = true WHERE id = $1 AND user_id = $2", id, userID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// In-memory fallback repository
type MemoryRepository struct {
	mu       sync.RWMutex
	vehicles map[string]*Vehicle
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		vehicles: make(map[string]*Vehicle),
	}
}

func (r *MemoryRepository) FindByUserID(ctx context.Context, userID string) ([]Vehicle, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []Vehicle
	for _, v := range r.vehicles {
		if v.UserID == userID {
			list = append(list, *v)
		}
	}
	return list, nil
}

func (r *MemoryRepository) Create(ctx context.Context, v *Vehicle) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	v.ID = fmt.Sprintf("veh_%d", time.Now().UnixNano())
	v.CreatedAt = time.Now().Format(time.RFC3339)
	v.UpdatedAt = time.Now().Format(time.RFC3339)

	userVehicles := 0
	for _, existing := range r.vehicles {
		if existing.UserID == v.UserID {
			userVehicles++
		}
	}
	if userVehicles == 0 {
		v.IsActive = true
	}

	r.vehicles[v.ID] = v
	return nil
}

func (r *MemoryRepository) Delete(ctx context.Context, id, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if v, ok := r.vehicles[id]; ok && v.UserID == userID {
		delete(r.vehicles, id)
	}
	return nil
}

func (r *MemoryRepository) SetActive(ctx context.Context, id, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, v := range r.vehicles {
		if v.UserID == userID {
			v.IsActive = (v.ID == id)
		}
	}
	return nil
}
