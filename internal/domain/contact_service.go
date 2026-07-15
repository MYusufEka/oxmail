package domain

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrContactNotFound = errors.New("contact not found")
	ErrContactExists   = errors.New("contact already exists")
)

// ContactService handles contact/address book CRUD operations.
type ContactService struct {
	conn *sql.DB
}

// NewContactService creates a ContactService with the given DB connection.
func NewContactService(conn *sql.DB) *ContactService {
	return &ContactService{conn: conn}
}

// Create adds a new contact for a user. Returns ErrContactExists if same user+email pair exists.
func (s *ContactService) Create(ctx context.Context, userEmail string, req CreateContactRequest) (*Contact, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}

	var existing int
	err := s.conn.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM contacts WHERE user_email = ? AND email = ?`,
		userEmail, req.Email,
	).Scan(&existing)
	if err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}
	if existing > 0 {
		return nil, ErrContactExists
	}

	now := time.Now().UTC()
	result, err := s.conn.ExecContext(ctx,
		`INSERT INTO contacts (user_email, name, email, phone, created_at) VALUES (?, ?, ?, ?, ?)`,
		userEmail, req.Name, req.Email, req.Phone, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert contact: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	return &Contact{
		ID:        id,
		UserEmail: userEmail,
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		CreatedAt: now,
	}, nil
}

// Get retrieves a single contact by ID.
func (s *ContactService) Get(ctx context.Context, id int64) (*Contact, error) {
	var c Contact
	err := s.conn.QueryRowContext(ctx,
		`SELECT id, user_email, name, email, phone, created_at FROM contacts WHERE id = ?`, id,
	).Scan(&c.ID, &c.UserEmail, &c.Name, &c.Email, &c.Phone, &c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrContactNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query contact: %w", err)
	}
	return &c, nil
}

// List returns all contacts for a given user email.
func (s *ContactService) List(ctx context.Context, userEmail string) ([]Contact, error) {
	rows, err := s.conn.QueryContext(ctx,
		`SELECT id, user_email, name, email, phone, created_at FROM contacts WHERE user_email = ? ORDER BY name ASC`,
		userEmail,
	)
	if err != nil {
		return nil, fmt.Errorf("list contacts: %w", err)
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var c Contact
		if err := rows.Scan(&c.ID, &c.UserEmail, &c.Name, &c.Email, &c.Phone, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan contact: %w", err)
		}
		contacts = append(contacts, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate contacts: %w", err)
	}

	if contacts == nil {
		contacts = []Contact{}
	}

	return contacts, nil
}

// Update modifies a contact. Only non-nil fields are changed.
func (s *ContactService) Update(ctx context.Context, id int64, req UpdateContactRequest) (*Contact, error) {
	existing, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	var sets []string
	var args []interface{}

	if req.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Email != nil {
		sets = append(sets, "email = ?")
		args = append(args, *req.Email)
	}
	if req.Phone != nil {
		sets = append(sets, "phone = ?")
		args = append(args, *req.Phone)
	}

	if len(sets) == 0 {
		return existing, nil
	}

	args = append(args, id)

	query := fmt.Sprintf("UPDATE contacts SET %s WHERE id = ?", joinComma(sets))
	_, err = s.conn.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("update contact: %w", err)
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Email != nil {
		existing.Email = *req.Email
	}
	if req.Phone != nil {
		existing.Phone = *req.Phone
	}

	return existing, nil
}

// Delete removes a contact by ID.
func (s *ContactService) Delete(ctx context.Context, id int64) error {
	result, err := s.conn.ExecContext(ctx, `DELETE FROM contacts WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete contact: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrContactNotFound
	}

	return nil
}

// ResolveName looks up a contact by user email + contact email and returns the display name.
// Returns empty string if no contact found.
func (s *ContactService) ResolveName(ctx context.Context, userEmail, contactEmail string) string {
	var name string
	err := s.conn.QueryRowContext(ctx,
		`SELECT name FROM contacts WHERE user_email = ? AND email = ? LIMIT 1`,
		userEmail, contactEmail,
	).Scan(&name)
	if err != nil {
		return ""
	}
	return name
}

// joinComma joins string slices with commas (replacement for strings.Join in dynamic SQL).
func joinComma(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}
