package domain

import (
	"database/sql"
	"errors"
	"fmt"
)

// AliasService handles alias CRUD operations and circular detection.
type AliasService struct {
	conn *sql.DB
}

// NewAliasService creates a new AliasService.
func NewAliasService(conn *sql.DB) *AliasService {
	return &AliasService{conn: conn}
}

// Create validates and inserts a new alias.
func (s *AliasService) Create(source, destination string) (*Alias, error) {
	if err := validateEmail(source); err != nil {
		return nil, fmt.Errorf("invalid source address: %w", err)
	}
	if err := validateEmail(destination); err != nil {
		return nil, fmt.Errorf("invalid destination address: %w", err)
	}

	sourceDomain := extractDomain(source)
	domainID, err := s.lookupDomainID(sourceDomain)
	if err != nil {
		return nil, err
	}

	if err := s.checkDuplicate(source, destination); err != nil {
		return nil, err
	}

	if err := s.detectCircular(source, destination); err != nil {
		return nil, err
	}

	result, err := s.conn.Exec(
		`INSERT INTO aliases (source_address, destination_address, domain_id, active) VALUES (?, ?, ?, 1)`,
		source, destination, domainID,
	)
	if err != nil {
		return nil, fmt.Errorf("insert alias: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	return s.Get(id)
}

// Get retrieves an alias by ID.
func (s *AliasService) Get(id int64) (*Alias, error) {
	alias := &Alias{}
	err := s.conn.QueryRow(
		`SELECT id, source_address, destination_address, active, created_at FROM aliases WHERE id = ?`,
		id,
	).Scan(&alias.ID, &alias.SourceAddress, &alias.DestinationAddress, &alias.Active, &alias.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("alias not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query alias: %w", err)
	}

	return alias, nil
}

// List returns all aliases.
func (s *AliasService) List() ([]Alias, error) {
	rows, err := s.conn.Query(
		`SELECT id, source_address, destination_address, active, created_at FROM aliases ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("query aliases: %w", err)
	}
	defer rows.Close()

	var aliases []Alias
	for rows.Next() {
		var alias Alias
		if err := rows.Scan(&alias.ID, &alias.SourceAddress, &alias.DestinationAddress, &alias.Active, &alias.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan alias: %w", err)
		}
		aliases = append(aliases, alias)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate aliases: %w", err)
	}

	if aliases == nil {
		aliases = []Alias{}
	}

	return aliases, nil
}

// Delete removes an alias by ID.
func (s *AliasService) Delete(id int64) error {
	result, err := s.conn.Exec(`DELETE FROM aliases WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete alias: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alias not found")
	}

	return nil
}

// GetAll returns all aliases (used by config generators).
func (s *AliasService) GetAll() ([]Alias, error) {
	return s.List()
}

func (s *AliasService) lookupDomainID(domainName string) (int64, error) {
	var id int64
	err := s.conn.QueryRow(`SELECT id FROM domains WHERE name = ? AND active = 1`, domainName).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrDomainNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("lookup domain: %w", err)
	}
	return id, nil
}

func (s *AliasService) checkDuplicate(source, destination string) error {
	var count int
	err := s.conn.QueryRow(
		`SELECT COUNT(*) FROM aliases WHERE source_address = ? AND destination_address = ?`,
		source, destination,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("check duplicate: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("alias already exists: %s -> %s", source, destination)
	}
	return nil
}

// detectCircular checks if adding source->destination would create a cycle.
func (s *AliasService) detectCircular(source, destination string) error {
	rows, err := s.conn.Query(`SELECT source_address, destination_address FROM aliases`)
	if err != nil {
		return fmt.Errorf("query aliases for cycle detection: %w", err)
	}
	defer rows.Close()

	graph := make(map[string][]string)
	for rows.Next() {
		var src, dst string
		if err := rows.Scan(&src, &dst); err != nil {
			return fmt.Errorf("scan alias for cycle detection: %w", err)
		}
		graph[src] = append(graph[src], dst)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate aliases for cycle detection: %w", err)
	}

	// Add the proposed edge.
	graph[source] = append(graph[source], destination)

	// BFS from destination to see if we can reach source.
	if reachable(graph, destination, source) {
		return fmt.Errorf("circular alias detected: %s -> %s creates a cycle", source, destination)
	}

	return nil
}

// reachable checks if target is reachable from start via the graph.
func reachable(graph map[string][]string, start, target string) bool {
	visited := make(map[string]bool)
	queue := []string{start}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == target {
			return true
		}

		if visited[current] {
			continue
		}
		visited[current] = true

		for _, neighbor := range graph[current] {
			if !visited[neighbor] {
				queue = append(queue, neighbor)
			}
		}
	}

	return false
}
