package storage

import (
	"context"
	"database/sql"
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

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS urls(id int, short text not null, original text not null)")
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return &RDB{
		db: db,
	}
}

func (d *RDB) Save(originalURL string, shortURL string) error {
	_, err := d.db.Exec("INSERT INTO urls(short, original) VALUES ($1, $2)", shortURL, originalURL)
	if err != nil {
		return err
	}
	return nil
}

func (d *RDB) Get(inputURL string) (string, error) {
	var short, original string

	row := d.db.QueryRowContext(context.Background(), "SELECT short, original FROM urls WHERE short = $1 or original = $1", inputURL)
	if err := row.Scan(&short, &original); err != nil {
		return "", err
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
