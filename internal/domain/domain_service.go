package domain

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/MYusufEka/oxmail/internal/database"
)

var (
	ErrDomainExists   = errors.New("domain already exists")
	ErrDomainNotFound = errors.New("domain not found")
	ErrInvalidDomain  = errors.New("invalid domain name")
)

// domainNameRegex validates RFC-compliant domain names.
// Each label: starts with alnum, can contain hyphens, ends with alnum, 1-63 chars.
// Must have at least two labels (name + TLD).
var domainNameRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// DomainService handles domain CRUD operations.
type DomainService struct {
	db *database.DB
}

// NewDomainService creates a new DomainService.
func NewDomainService(db *database.DB) *DomainService {
	return &DomainService{db: db}
}

// Create adds a new domain after validation.
func (s *DomainService) Create(ctx context.Context, name string) (*Domain, error) {
	if err := validateDomainName(name); err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	result, err := s.db.Conn.ExecContext(ctx,
		"INSERT INTO domains (name, active, created_at, updated_at) VALUES (?, 1, ?, ?)",
		name, now, now,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrDomainExists
		}
		return nil, fmt.Errorf("insert domain: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	return &Domain{
		ID:        id,
		Name:      name,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Get retrieves a domain by name.
func (s *DomainService) Get(ctx context.Context, name string) (*Domain, error) {
	var d Domain
	err := s.db.Conn.QueryRowContext(ctx,
		"SELECT id, name, active, created_at, updated_at FROM domains WHERE name = ?",
		name,
	).Scan(&d.ID, &d.Name, &d.Active, &d.CreatedAt, &d.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrDomainNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query domain: %w", err)
	}

	return &d, nil
}

// GetDomainByName implements DomainLookup. Returns nil, nil if not found.
func (s *DomainService) GetDomainByName(ctx context.Context, name string) (*Domain, error) {
	d, err := s.Get(ctx, name)
	if errors.Is(err, ErrDomainNotFound) {
		return nil, nil
	}
	return d, err
}

func (s *DomainService) GetByID(ctx context.Context, id int64) (string, error) {
	var name string
	err := s.db.Conn.QueryRowContext(ctx,
		"SELECT name FROM domains WHERE id = ?",
		id,
	).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrDomainNotFound
	}
	if err != nil {
		return "", fmt.Errorf("query domain by id: %w", err)
	}
	return name, nil
}

// List returns a paginated list of domains and the total count.
func (s *DomainService) List(ctx context.Context, page, limit int) ([]Domain, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	var total int
	err := s.db.Conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM domains").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count domains: %w", err)
	}

	offset := (page - 1) * limit
	rows, err := s.db.Conn.QueryContext(ctx,
		"SELECT id, name, active, created_at, updated_at FROM domains ORDER BY id ASC LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list domains: %w", err)
	}
	defer rows.Close()

	var domains []Domain
	for rows.Next() {
		var d Domain
		if err := rows.Scan(&d.ID, &d.Name, &d.Active, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan domain: %w", err)
		}
		domains = append(domains, d)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate domains: %w", err)
	}

	return domains, total, nil
}

// Delete removes a domain by name.
func (s *DomainService) Delete(ctx context.Context, name string) error {
	result, err := s.db.Conn.ExecContext(ctx, "DELETE FROM domains WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("delete domain: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrDomainNotFound
	}

	return nil
}

// validateDomainName checks if a domain name is RFC-compliant.
func validateDomainName(name string) error {
	if name == "" {
		return ErrInvalidDomain
	}
	if strings.Contains(name, " ") {
		return ErrInvalidDomain
	}
	if strings.HasSuffix(name, ".") {
		return ErrInvalidDomain
	}
	if !domainNameRegex.MatchString(name) {
		return ErrInvalidDomain
	}
	if len(name) > 253 {
		return ErrInvalidDomain
	}
	return nil
}
