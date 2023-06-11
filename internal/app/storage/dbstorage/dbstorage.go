package dbstorage

import (
	"context"
	"database/sql"
	"fmt"
	"shorty/internal/app/hash"
	"shorty/internal/app/models"
	"time"
)

type storage struct {
	db *sql.DB
}

func CreateDBStorage(databaseDSN string) (*storage, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}

	if err := setUpDatabase(db); err != nil {
		return nil, err
	}

	s := &storage{db: db}

	return s, nil
}

func (s *storage) Get(key string) (string, error) {
	conn, err := s.db.Conn(context.TODO())
	if err != nil {
		return "", fmt.Errorf("failed to open connection to db: %v", err)
	}
	defer conn.Close()
	row := conn.QueryRowContext(context.TODO(), "SELECT url FROM links WHERE hash = $1", key)
	var url string
	if err := row.Scan(&url); err != nil {
		return "", fmt.Errorf("failed to scan row: %v", err)
	}
	return url, nil
}

func (s *storage) Put(key, value string) error {
	conn, err := s.db.Conn(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to open connection to db: %v", err)
	}

	_, err = conn.ExecContext(context.TODO(), "INSERT INTO links (hash, url) VALUES ($1, $2)", key, value)
	if err != nil {
		return fmt.Errorf("could not insert row: %v", err)
	}

	return nil
}

func (s *storage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := s.db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

func (s *storage) Batch(urls models.ShortenBatchRequest) error {
	fmt.Println("db storage batching!!!", urls)
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	for _, url := range urls {
		_, err := tx.ExecContext(context.TODO(), "INSERT INTO links (hash, url) VALUES ($1, $2)", hash.Generate([]byte(url.OriginalURL)), url.OriginalURL)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert line in table with url=%s: %v", url.OriginalURL, err)
		}
	}

	return tx.Commit()
}

func setUpDatabase(db *sql.DB) error {
	conn, err := db.Conn(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to open connection to db: %v", err)
	}
	defer conn.Close()

	_, err = conn.ExecContext(context.TODO(), `CREATE TABLE IF NOT EXISTS links (id SERIAL PRIMARY KEY, hash VARCHAR(8), url VARCHAR(1024))`)

	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)

	}
	return nil
}
