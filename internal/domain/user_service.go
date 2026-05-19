package domain

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MYusufEka/oxmail/internal/database"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrInvalidEmail = errors.New("invalid email address")
	ErrUserNotFound = errors.New("user not found")
)

// DomainLookup provides domain resolution without coupling to DomainService directly.
type DomainLookup interface {
	GetDomainByName(ctx context.Context, name string) (*Domain, error)
}

// UserListParams holds filtering and pagination for listing users.
type UserListParams struct {
	Domain string
	Page   int
	Limit  int
}

// UserService handles user/mailbox CRUD operations.
type UserService struct {
	db           *database.DB
	domainLookup DomainLookup
}

// NewUserService creates a UserService with the given dependencies.
func NewUserService(db *database.DB, domainLookup DomainLookup) *UserService {
	return &UserService{
		db:           db,
		domainLookup: domainLookup,
	}
}

// Create validates input, hashes the password, and inserts a new user.
func (s *UserService) Create(ctx context.Context, req CreateUserRequest) (*User, error) {
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}

	domainName := extractDomain(req.Email)
	dom, err := s.domainLookup.GetDomainByName(ctx, domainName)
	if err != nil {
		return nil, fmt.Errorf("lookup domain: %w", err)
	}
	if dom == nil {
		return nil, ErrDomainNotFound
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	result, err := s.db.Conn.ExecContext(ctx,
		`INSERT INTO users (email, password_hash, domain_id, display_name, quota, active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, 1, ?, ?)`,
		req.Email, string(hash), dom.ID, req.DisplayName, req.Quota, now, now,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrUserExists
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	return &User{
		ID:           id,
		Email:        req.Email,
		PasswordHash: string(hash),
		DomainID:     dom.ID,
		DisplayName:  req.DisplayName,
		Quota:        req.Quota,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// GetByID retrieves a user by their ID.
func (s *UserService) GetByID(ctx context.Context, id int64) (*User, error) {
	var user User
	err := s.db.Conn.QueryRowContext(ctx,
		`SELECT id, email, password_hash, domain_id, display_name, quota, active, created_at, updated_at
		 FROM users WHERE id = ?`, id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DomainID,
		&user.DisplayName, &user.Quota, &user.Active, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}
	return &user, nil
}

// List returns users with optional domain filter and pagination.
func (s *UserService) List(ctx context.Context, params UserListParams) ([]User, int, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 20
	}

	var countQuery, listQuery strings.Builder
	var args []interface{}

	countQuery.WriteString("SELECT COUNT(*) FROM users")
	listQuery.WriteString("SELECT users.id, users.email, users.password_hash, users.domain_id, users.display_name, users.quota, users.active, users.created_at, users.updated_at FROM users")

	if params.Domain != "" {
		countQuery.WriteString(" JOIN domains ON users.domain_id = domains.id WHERE domains.name = ?")
		listQuery.WriteString(" JOIN domains ON users.domain_id = domains.id WHERE domains.name = ?")
		args = append(args, params.Domain)
	}

	var total int
	err := s.db.Conn.QueryRowContext(ctx, countQuery.String(), args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	offset := (params.Page - 1) * params.Limit
	listQuery.WriteString(" ORDER BY users.id ASC LIMIT ? OFFSET ?")
	listArgs := append(args, params.Limit, offset)

	rows, err := s.db.Conn.QueryContext(ctx, listQuery.String(), listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DomainID,
			&user.DisplayName, &user.Quota, &user.Active, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate users: %w", err)
	}

	return users, total, nil
}

// Delete removes a user by ID.
func (s *UserService) Delete(ctx context.Context, id int64) error {
	result, err := s.db.Conn.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

// GetByEmail retrieves a user by email address.
func (s *UserService) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := s.db.Conn.QueryRowContext(ctx,
		`SELECT id, email, password_hash, domain_id, display_name, quota, active, created_at, updated_at
		 FROM users WHERE email = ?`, email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DomainID,
		&user.DisplayName, &user.Quota, &user.Active, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query user by email: %w", err)
	}
	return &user, nil
}
