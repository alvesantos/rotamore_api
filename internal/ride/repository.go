package ride

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	FindByUserID(ctx context.Context, userID string, limit int, year int) ([]Ride, error)
	FindByID(ctx context.Context, id, userID string) (*Ride, error)
	Create(ctx context.Context, r *Ride) error
	Delete(ctx context.Context, id, userID string) error
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) FindByUserID(ctx context.Context, userID string, limit int, year int) ([]Ride, error) {
	var query string
	var rowsArgs []interface{}

	if year > 0 {
		query = `
			SELECT 
				r.id, r.user_id, r.vehicle_id,
				COALESCE(v.name, ''), COALESCE(v.brand, ''), COALESCE(v.plate, ''), COALESCE(v.color, ''),
				r.customer_name, r.customer_phone, r.passengers_count,
				r.pickup, r.destination, COALESCE(r.notes, ''),
				to_char(r.ride_date, 'YYYY-MM-DD'), r.ride_time, r.price, r.status,
				r.created_at, r.updated_at
			FROM rides r
			LEFT JOIN vehicles v ON r.vehicle_id = v.id
			WHERE r.user_id = $1 AND EXTRACT(YEAR FROM r.ride_date) = $2
			ORDER BY r.ride_date DESC, r.ride_time DESC
		`
		rowsArgs = []interface{}{userID, year}
	} else if limit > 0 {
		query = `
			SELECT 
				r.id, r.user_id, r.vehicle_id,
				COALESCE(v.name, ''), COALESCE(v.brand, ''), COALESCE(v.plate, ''), COALESCE(v.color, ''),
				r.customer_name, r.customer_phone, r.passengers_count,
				r.pickup, r.destination, COALESCE(r.notes, ''),
				to_char(r.ride_date, 'YYYY-MM-DD'), r.ride_time, r.price, r.status,
				r.created_at, r.updated_at
			FROM rides r
			LEFT JOIN vehicles v ON r.vehicle_id = v.id
			WHERE r.user_id = $1
			ORDER BY r.ride_date DESC, r.ride_time DESC
			LIMIT $2
		`
		rowsArgs = []interface{}{userID, limit}
	} else {
		query = `
			SELECT 
				r.id, r.user_id, r.vehicle_id,
				COALESCE(v.name, ''), COALESCE(v.brand, ''), COALESCE(v.plate, ''), COALESCE(v.color, ''),
				r.customer_name, r.customer_phone, r.passengers_count,
				r.pickup, r.destination, COALESCE(r.notes, ''),
				to_char(r.ride_date, 'YYYY-MM-DD'), r.ride_time, r.price, r.status,
				r.created_at, r.updated_at
			FROM rides r
			LEFT JOIN vehicles v ON r.vehicle_id = v.id
			WHERE r.user_id = $1
			ORDER BY r.ride_date DESC, r.ride_time DESC
		`
		rowsArgs = []interface{}{userID}
	}

	rows, err := r.pool.Query(ctx, query, rowsArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rides []Ride
	for rows.Next() {
		var item Ride
		var vehicleID sql.NullString
		var createdAt, updatedAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&vehicleID,
			&item.VehicleName,
			&item.VehicleBrand,
			&item.VehiclePlate,
			&item.VehicleColor,
			&item.CustomerName,
			&item.CustomerPhone,
			&item.PassengersCount,
			&item.Pickup,
			&item.Destination,
			&item.Notes,
			&item.RideDate,
			&item.RideTime,
			&item.Price,
			&item.Status,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}

		if vehicleID.Valid {
			vID := vehicleID.String
			item.VehicleID = &vID
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		rides = append(rides, item)
	}

	return rides, nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id, userID string) (*Ride, error) {
	query := `
		SELECT 
			r.id, r.user_id, r.vehicle_id,
			COALESCE(v.name, ''), COALESCE(v.brand, ''), COALESCE(v.plate, ''), COALESCE(v.color, ''),
			r.customer_name, r.customer_phone, r.passengers_count,
			r.pickup, r.destination, COALESCE(r.notes, ''),
			to_char(r.ride_date, 'YYYY-MM-DD'), r.ride_time, r.price, r.status,
			r.created_at, r.updated_at
		FROM rides r
		LEFT JOIN vehicles v ON r.vehicle_id = v.id
		WHERE r.id = $1 AND r.user_id = $2
		LIMIT 1
	`

	var item Ride
	var vehicleID sql.NullString
	var createdAt, updatedAt time.Time
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&item.ID,
		&item.UserID,
		&vehicleID,
		&item.VehicleName,
		&item.VehicleBrand,
		&item.VehiclePlate,
		&item.VehicleColor,
		&item.CustomerName,
		&item.CustomerPhone,
		&item.PassengersCount,
		&item.Pickup,
		&item.Destination,
		&item.Notes,
		&item.RideDate,
		&item.RideTime,
		&item.Price,
		&item.Status,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	if vehicleID.Valid {
		vID := vehicleID.String
		item.VehicleID = &vID
	}
	item.CreatedAt = createdAt.Format(time.RFC3339)
	item.UpdatedAt = updatedAt.Format(time.RFC3339)
	return &item, nil
}

func (r *PostgresRepository) Create(ctx context.Context, item *Ride) error {
	if item.Status == "" {
		item.Status = "agendada"
	}

	query := `
		INSERT INTO rides (
			user_id, vehicle_id, customer_name, customer_phone, passengers_count,
			pickup, destination, notes, ride_date, ride_time, price, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::DATE, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err := r.pool.QueryRow(
		ctx,
		query,
		item.UserID,
		item.VehicleID,
		item.CustomerName,
		item.CustomerPhone,
		item.PassengersCount,
		item.Pickup,
		item.Destination,
		item.Notes,
		item.RideDate,
		item.RideTime,
		item.Price,
		item.Status,
	).Scan(&item.ID, &createdAt, &updatedAt)

	if err != nil {
		return fmt.Errorf("erro ao salvar corrida: %w", err)
	}

	// Retrieve vehicle info if vehicle_id was provided
	if item.VehicleID != nil && *item.VehicleID != "" {
		_ = r.pool.QueryRow(
			ctx,
			"SELECT name, brand, plate, color FROM vehicles WHERE id = $1",
			*item.VehicleID,
		).Scan(&item.VehicleName, &item.VehicleBrand, &item.VehiclePlate, &item.VehicleColor)
	}

	item.CreatedAt = createdAt.Format(time.RFC3339)
	item.UpdatedAt = updatedAt.Format(time.RFC3339)
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id, userID string) error {
	query := `DELETE FROM rides WHERE id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, id, userID)
	return err
}

// Memory Repository for fallbacks and unit testing
type MemoryRepository struct {
	mu    sync.RWMutex
	rides map[string]*Ride
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		rides: make(map[string]*Ride),
	}
}

func (r *MemoryRepository) FindByUserID(ctx context.Context, userID string, limit int, year int) ([]Ride, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []Ride
	for _, item := range r.rides {
		if item.UserID == userID {
			list = append(list, *item)
		}
	}
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

func (r *MemoryRepository) FindByID(ctx context.Context, id, userID string) (*Ride, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if item, ok := r.rides[id]; ok && item.UserID == userID {
		return item, nil
	}
	return nil, fmt.Errorf("corrida não encontrada")
}

func (r *MemoryRepository) Create(ctx context.Context, item *Ride) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item.ID = fmt.Sprintf("ride_%d", time.Now().UnixNano())
	if item.Status == "" {
		item.Status = "agendada"
	}
	item.CreatedAt = time.Now().Format(time.RFC3339)
	item.UpdatedAt = time.Now().Format(time.RFC3339)
	r.rides[item.ID] = item
	return nil
}

func (r *MemoryRepository) Delete(ctx context.Context, id, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if item, ok := r.rides[id]; ok && item.UserID == userID {
		delete(r.rides, id)
	}
	return nil
}
