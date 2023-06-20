package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"shorty/internal/app/hash"
	"shorty/internal/app/models"
	"strconv"
	"strings"
	"time"
)

var ErrConflict = errors.New("url already saved")

type dbstorage struct {
	db *sql.DB
}

func CreateDBStorage(ctx context.Context, databaseDSN string) (*dbstorage, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection to database with %s, %v", databaseDSN, err)
	}

	if err := setUpDatabase(ctx, db); err != nil {
		return nil, err
	}

	s := &dbstorage{db: db}

	return s, nil
}

func (s *dbstorage) Get(ctx context.Context, key string) (string, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to open connection to db: %v", err)
	}
	defer conn.Close()
	row := conn.QueryRowContext(ctx, "SELECT url FROM links WHERE hash = $1", key)
	var url string
	if err := row.Scan(&url); err != nil {
		return "", fmt.Errorf("failed to scan row: %v", err)
	}
	return url, nil
}

func (s *dbstorage) Put(ctx context.Context, key, value, userID string) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to open connection to db: %w", err)
	}
	defer conn.Close()

	result, err := conn.ExecContext(ctx, "INSERT INTO links (hash, url, user_id) VALUES ($1, $2, $3) ON CONFLICT (url) DO NOTHING", key, value, userID)
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

func (s *dbstorage) Batch(ctx context.Context, urls models.ShortenBatchRequest, userID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	values := make([]string, len(urls))

	for index := range urls {
		values[index] = "( $" + strconv.Itoa(2*index+1) + ", $" + strconv.Itoa(2*index+2) + ", $" + strconv.Itoa(2*len(urls)+1) + ")"
	}

	query := "INSERT INTO links (hash, url, user_id) VALUES " + strings.Join(values, ", ")

	_, err = tx.ExecContext(ctx, query, generateQueryKeys(urls, userID)...)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert line in table with %v", err)
	}

	return tx.Commit()
}

func (s *dbstorage) UserURLs(ctx context.Context, userID string) ([]models.StorageURLsTODO, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection to db: %w", err)
	}
	defer conn.Close()

	rows, err := conn.QueryContext(ctx, "SELECT hash, url FROM links WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users rows with userID=%s: %v", userID, err)
	}
	defer rows.Close()

	var urls []models.StorageURLsTODO
	for rows.Next() {
		var row models.StorageURLsTODO

		err = rows.Scan(&row.Hash, &row.URL)
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

	_, err = conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS links (id SERIAL PRIMARY KEY, hash VARCHAR(8), url VARCHAR(1024), user_id VARCHAR(36))`)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	_, err = conn.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS id_url ON links (url)`)
	if err != nil {
		return fmt.Errorf("failed to set index: %v", err)
	}
	return nil
}

func generateQueryKeys(urls models.ShortenBatchRequest, userID string) []any {
	keys := []any{}
	for _, row := range urls {
		keys = append(keys, hash.Generate([]byte(row.OriginalURL)), row.OriginalURL)
	}
	keys = append(keys, userID)
	return keys
}
