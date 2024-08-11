package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
)

type RDB struct {
	db *sql.DB
}

func NewDB(DBPath string) *RDB {
	db, err := sql.Open("pgx", DBPath)
	if err != nil {
		log.Fatal(err)
	}

	createUrlsTableQuery := `CREATE TABLE IF NOT EXISTS urls(
		id serial primary key,
		short_url varchar(255) not null,
		original_url varchar(255) not null
		)`
	_, err = db.Exec(createUrlsTableQuery)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return &RDB{
		db: db,
	}
}

func (d *RDB) Save(originalURL string, shortURL string) error {
	_, err := d.db.Exec("INSERT INTO urls(short_url, original_url) VALUES ($1, $2)", shortURL, originalURL)
	if err != nil {
		return err
	}
	return nil
}

func (d *RDB) Get(inputURL string) (string, error) {
	var short, original string

	row := d.db.QueryRowContext(context.Background(), "SELECT short_url, original_url FROM urls WHERE short_url = $1 or original_url = $1", inputURL)
	if err := row.Scan(&short, &original); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("cannot scan row: %w", err)
	}

	if short != "" {
		return short, nil
	}

	return original, nil
}

func (d *RDB) Ping(ctx context.Context) error {
	if err := d.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}
