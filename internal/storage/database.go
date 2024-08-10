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

	return &RDB{
		db: db,
	}
}

func (d *RDB) Save(originalURL string, shortURL string) error {
	panic("No implement")
}

func (d *RDB) Get(inputURL string) (string, error) {
	panic("No implement")
}

func (d *RDB) Ping(ctx context.Context) error {
	if err := d.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}
