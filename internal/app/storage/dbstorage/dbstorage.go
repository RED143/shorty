package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"shorty/internal/app/config"
	"shorty/internal/app/models"
	"strconv"
	"strings"
	"time"
)

var ErrConflict = errors.New("url already saved")

type dbstorage struct {
	db *sql.DB
}

type GetURL struct {
	originalURL string
	isDeleted   bool
}

func CreateDBStorage(ctx context.Context, cfg config.Config) (*dbstorage, error) {
	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	db.SetMaxOpenConns(cfg.MaxDBConnections)
	db.SetMaxIdleConns(cfg.MaxDBConnections)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection to database with %s, %v", cfg.DatabaseDSN, err)
	}

	if err := setUpDatabase(ctx, db); err != nil {
		return nil, err
	}

	s := &dbstorage{db: db}

	return s, nil
}

func (s *dbstorage) Get(ctx context.Context, shortURL string) (models.UserURLs, error) {
	var urls models.UserURLs
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return urls, fmt.Errorf("failed to open connection to db: %v", err)
	}
	defer conn.Close()
	row := conn.QueryRowContext(ctx, "SELECT short_url, original_url, is_deleted FROM links WHERE short_url = $1", shortURL)

	if err := row.Scan(&urls.ShortURL, &urls.OriginalURL, &urls.IsDeleted); err != nil {
		return urls, fmt.Errorf("failed to scan row: %v", err)
	}
	return urls, nil
}

func (s *dbstorage) Put(ctx context.Context, shortURL, originalURL, userID string) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to open connection to db: %w", err)
	}
	defer conn.Close()

	result, err := conn.ExecContext(ctx, "INSERT INTO links (short_url, original_url, user_id) VALUES ($1, $2, $3) ON CONFLICT (original_url) DO NOTHING", shortURL, originalURL, userID)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to insert row: %v", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to count affected rows %v", err)
	}

	if count == 0 {
		conn.Close()
		return ErrConflict
	}

	return nil
}

func (s *dbstorage) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	if err := s.db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

func (s *dbstorage) Batch(ctx context.Context, urls []models.UserURLs, userID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	values := make([]string, len(urls))

	for index := range urls {
		values[index] = "( $" + strconv.Itoa(2*index+1) + ", $" + strconv.Itoa(2*index+2) + ", $" + strconv.Itoa(2*len(urls)+1) + ")"
	}

	query := "INSERT INTO links (short_url, original_url, user_id) VALUES " + strings.Join(values, ", ")

	_, err = tx.ExecContext(ctx, query, generateQueryValues(urls, userID)...)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert line in table with %v", err)
	}

	return tx.Commit()
}

func (s *dbstorage) UserURLs(ctx context.Context, userID string) ([]models.UserURLs, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection to db: %w", err)
	}
	defer conn.Close()

	rows, err := conn.QueryContext(ctx, "SELECT short_url, original_url FROM links WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users rows with userID=%s: %v", userID, err)
	}
	defer rows.Close()

	var urls []models.UserURLs
	for rows.Next() {
		var row models.UserURLs

		err = rows.Scan(&row.ShortURL, &row.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("failed to scan urls %v", err)
		}

		urls = append(urls, row)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("failed during selecting urls %v", err)
	}

	return urls, nil
}

func (s *dbstorage) DeleteUserURls(ctx context.Context, urls []string, userID string) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to open connection to db: %w", err)
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "UPDATE links SET is_deleted = true WHERE user_id = $1 AND short_url = any($2)", userID, urls)
	if err != nil {
		return fmt.Errorf("failed to update is_deleted column: %w", err)
	}

	return nil
}

func (s *dbstorage) Close() error {
	err := s.db.Close()
	if err != nil {
		return fmt.Errorf("failed to close db: %v", err)
	}
	return nil
}

func setUpDatabase(ctx context.Context, db *sql.DB) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to open connection to db: %v", err)
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS links (id SERIAL PRIMARY KEY, short_url VARCHAR(128), original_url VARCHAR(1024), user_id VARCHAR(36), is_deleted BOOLEAN DEFAULT false)`)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	_, err = conn.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS id_url ON links (original_url)`)
	if err != nil {
		return fmt.Errorf("failed to set index: %v", err)
	}
	return nil
}

func generateQueryValues(urls []models.UserURLs, userID string) []any {
	keys := []any{}
	for _, row := range urls {
		keys = append(keys, row.ShortURL, row.OriginalURL)
	}
	keys = append(keys, userID)
	return keys
}
