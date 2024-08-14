package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/romanp1989/go-shortener/internal/models"
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
		original_url varchar(255) not null);
                               
	    CREATE UNIQUE INDEX IF NOT EXISTS original_url_idx ON urls (original_url);`
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

	if inputURL == short {
		return original, nil
	}

	if inputURL == original {
		return short, nil
	}

	return "", nil
}

func (d *RDB) SaveBatch(ctx context.Context, urls []models.StorageURL) ([]string, error) {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var args []interface{}
	var insertValues string
	var shortURLs []string

	var paramNumber = 0
	for i, url := range urls {
		if i > 0 {
			insertValues += ","
		}
		insertValues += fmt.Sprintf("($%d, $%d)", paramNumber+1, paramNumber+2)

		args = append(args, url.ShortURL)
		args = append(args, url.OriginalURL)
		paramNumber += 2
	}

	query := `INSERT INTO urls (short_url, original_url) 
			 	VALUES ` + insertValues + `
				ON CONFLICT (original_url) DO UPDATE SET short_url = EXCLUDED.short_url, original_url = EXCLUDED.original_url
				RETURNING short_url`

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка при вставке записей: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var short string
		if err := rows.Scan(&short); err != nil {
			return nil, err
		}
		shortURLs = append(shortURLs, short)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(urls) != len(shortURLs) {
		return nil, errors.New("количество url в запросе не совпадает с числом сохраненных")
	}

	tx.Commit()

	return shortURLs, nil
}

func (d *RDB) Ping(ctx context.Context) error {
	if err := d.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}
