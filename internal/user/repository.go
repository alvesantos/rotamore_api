package user

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("usuário não encontrado")
	ErrInvalidPassword   = errors.New("senha incorreta")
	ErrUserAlreadyExists = errors.New("usuário já cadastrado com este e-mail ou documento")
)

type Repository interface {
	FindByEmailOrPhone(ctx context.Context, identifier string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	Update(ctx context.Context, u *User) error
	SeedDefaultUsers(ctx context.Context) error
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) FindByEmailOrPhone(ctx context.Context, identifier string) (*User, error) {
	identifier = strings.TrimSpace(identifier)
	cleanPhone := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(identifier, "(", ""), ")", ""), "-", ""), " ", "")

	query := `
		SELECT id, name, lastname, phone, type, email, document, COALESCE(state, 'AL'), password_hash, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1) OR phone = $1 OR phone = $2
		LIMIT 1
	`

	var u User
	var createdAt, updatedAt time.Time
	err := r.pool.QueryRow(ctx, query, identifier, cleanPhone).Scan(
		&u.ID,
		&u.Name,
		&u.LastName,
		&u.Phone,
		&u.Type,
		&u.Email,
		&u.Document,
		&u.State,
		&u.PasswordHash,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, ErrUserNotFound
	}

	u.CreatedAt = createdAt.Format(time.RFC3339)
	u.UpdatedAt = updatedAt.Format(time.RFC3339)
	return &u, nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, name, lastname, phone, type, email, document, COALESCE(state, 'AL'), password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`

	var u User
	var createdAt, updatedAt time.Time
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Name,
		&u.LastName,
		&u.Phone,
		&u.Type,
		&u.Email,
		&u.Document,
		&u.State,
		&u.PasswordHash,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, ErrUserNotFound
	}

	u.CreatedAt = createdAt.Format(time.RFC3339)
	u.UpdatedAt = updatedAt.Format(time.RFC3339)
	return &u, nil
}

func (r *PostgresRepository) Update(ctx context.Context, u *User) error {
	query := `
		UPDATE users
		SET name = $1, lastname = $2, phone = $3, email = $4, document = $5, state = $6, updated_at = now()
		WHERE id = $7
	`
	tag, err := r.pool.Exec(ctx, query, u.Name, u.LastName, u.Phone, u.Email, u.Document, u.State, u.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *PostgresRepository) SeedDefaultUsers(ctx context.Context) error {
	// Ensure table exists
	tableQuery := `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) NOT NULL,
			lastname VARCHAR(100) NOT NULL,
			phone VARCHAR(20) NOT NULL,
			type VARCHAR(20) NOT NULL CHECK (type IN ('driver', 'customer', 'admin')),
			email VARCHAR(255) NOT NULL UNIQUE,
			document VARCHAR(11) NOT NULL UNIQUE,
			state VARCHAR(50) DEFAULT 'AL',
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
	`
	if _, err := r.pool.Exec(ctx, tableQuery); err != nil {
		log.Printf("Aviso ao verificar tabela users: %v", err)
	}

	adminHash, _ := HashPassword("r0g4b@2026!")
	driverHash, _ := HashPassword("1254101254@Abc")

	seedQuery := `
		INSERT INTO users (name, lastname, phone, type, email, document, state, password_hash)
		VALUES 
			('Rogab', 'Admin', '11999999999', 'admin', 'rogab@admin.com', '00000000000', 'AL', $1),
			('Ricardo', 'Berns', '11988888888', 'driver', 'ricberns@gmail.com', '11111111111', 'AL', $2)
		ON CONFLICT (email) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			name = EXCLUDED.name,
			lastname = EXCLUDED.lastname,
			phone = EXCLUDED.phone,
			type = EXCLUDED.type,
			state = EXCLUDED.state,
			updated_at = now();
	`

	_, err := r.pool.Exec(ctx, seedQuery, adminHash, driverHash)
	if err != nil {
		return fmt.Errorf("erro ao executar seed: %w", err)
	}

	log.Println("✓ Seed de usuários padrão executado com sucesso no PostgreSQL!")
	return nil
}

// In-memory mock repository for fallback if DB is not reachable
type MemoryRepository struct {
	mu    sync.RWMutex
	users map[string]*User
}

func NewMemoryRepository() *MemoryRepository {
	repo := &MemoryRepository{
		users: make(map[string]*User),
	}
	_ = repo.SeedDefaultUsers(context.Background())
	return repo
}

func (r *MemoryRepository) SeedDefaultUsers(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	adminHash, _ := HashPassword("r0g4b@2026!")
	driverHash, _ := HashPassword("1254101254@Abc")

	admin := &User{
		ID:           "a0000000-0000-0000-0000-000000000001",
		Name:         "Rogab",
		LastName:     "Admin",
		Phone:        "11999999999",
		Type:         TypeAdmin,
		Email:        "rogab@admin.com",
		Document:     "00000000000",
		State:        "AL",
		PasswordHash: adminHash,
		CreatedAt:    time.Now().Format(time.RFC3339),
		UpdatedAt:    time.Now().Format(time.RFC3339),
	}

	driver := &User{
		ID:           "d0000000-0000-0000-0000-000000000002",
		Name:         "Ricardo",
		LastName:     "Berns",
		Phone:        "11988888888",
		Type:         TypeDriver,
		Email:        "ricberns@gmail.com",
		Document:     "11111111111",
		State:        "AL",
		PasswordHash: driverHash,
		CreatedAt:    time.Now().Format(time.RFC3339),
		UpdatedAt:    time.Now().Format(time.RFC3339),
	}

	r.users[admin.Email] = admin
	r.users[admin.Phone] = admin
	r.users[admin.ID] = admin

	r.users[driver.Email] = driver
	r.users[driver.Phone] = driver
	r.users[driver.ID] = driver

	log.Println("✓ Seed de usuários carregado em memória (Fallback)")
	return nil
}

func (r *MemoryRepository) FindByEmailOrPhone(ctx context.Context, identifier string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	identifier = strings.TrimSpace(identifier)
	cleanPhone := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(identifier, "(", ""), ")", ""), "-", ""), " ", "")

	for _, u := range r.users {
		if strings.EqualFold(u.Email, identifier) || u.Phone == identifier || u.Phone == cleanPhone {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (r *MemoryRepository) FindByID(ctx context.Context, id string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, ErrUserNotFound
}

func (r *MemoryRepository) Update(ctx context.Context, u *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.users[u.ID]; ok {
		existing.Name = u.Name
		existing.LastName = u.LastName
		existing.Phone = u.Phone
		existing.Email = u.Email
		existing.Document = u.Document
		existing.State = u.State
		existing.UpdatedAt = time.Now().Format(time.RFC3339)
		return nil
	}
	return ErrUserNotFound
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
