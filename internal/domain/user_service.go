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
		`SELECT id, email, password_hash, domain_id, display_name, quota, active, must_change_password, created_at, updated_at
		 FROM users WHERE id = ?`, id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DomainID,
		&user.DisplayName, &user.Quota, &user.Active, &user.MustChangePassword, &user.CreatedAt, &user.UpdatedAt)

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
	listQuery.WriteString("SELECT users.id, users.email, users.password_hash, users.domain_id, users.display_name, users.quota, users.active, users.must_change_password, users.created_at, users.updated_at FROM users")

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
			&user.DisplayName, &user.Quota, &user.Active, &user.MustChangePassword, &user.CreatedAt, &user.UpdatedAt); err != nil {
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

// Update partially updates a user. Only non-nil fields are changed.
// Password is re-hashed with bcrypt cost 12 only if non-empty.
func (s *UserService) Update(ctx context.Context, id int64, req UpdateUserRequest) (*User, error) {
	user, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var sets []string
	var args []interface{}
	var passwordHash string

	if req.Password != nil && *req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), 12)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		passwordHash = string(hash)
		sets = append(sets, "password_hash = ?", "must_change_password = 0")
		args = append(args, passwordHash)
	}

	if req.DisplayName != nil {
		sets = append(sets, "display_name = ?")
		args = append(args, *req.DisplayName)
	}

	if req.Quota != nil {
		sets = append(sets, "quota = ?")
		args = append(args, *req.Quota)
	}

	if req.MustChangePassword != nil {
		sets = append(sets, "must_change_password = ?")
		args = append(args, *req.MustChangePassword)
	}

	if len(sets) == 0 {
		return user, nil
	}

	now := time.Now().UTC()
	sets = append(sets, "updated_at = ?")
	args = append(args, now)
	args = append(args, id)

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(sets, ", "))
	_, err = s.db.Conn.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	user.UpdatedAt = now
	if req.Password != nil && *req.Password != "" {
		user.PasswordHash = passwordHash
		user.MustChangePassword = false
	}
	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.Quota != nil {
		user.Quota = *req.Quota
	}
	if req.MustChangePassword != nil {
		user.MustChangePassword = *req.MustChangePassword
	}

	return user, nil
}

// GetByEmail retrieves a user by email address.
func (s *UserService) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := s.db.Conn.QueryRowContext(ctx,
		`SELECT id, email, password_hash, domain_id, display_name, quota, active, must_change_password, created_at, updated_at
		 FROM users WHERE email = ?`, email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DomainID,
		&user.DisplayName, &user.Quota, &user.Active, &user.MustChangePassword, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query user by email: %w", err)
	}
	return &user, nil
}
