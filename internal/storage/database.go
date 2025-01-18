package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/romanp1989/go-shortener/internal/models"
	"log"
	"sync"
)

// SQLDB database operations interface
type SQLDB interface {
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	PingContext(ctx context.Context) error
}

// DBStorage DB storage
type DBStorage struct {
	db SQLDB
	mu sync.RWMutex
}

// SaveInsertQuery insert query for save urls
const SaveInsertQuery = `INSERT INTO urls(short_url, original_url, user_id) 
VALUES ($1, $2, $3)
RETURNING short_url`

// GetSelectQuery get url by short or original url
const GetSelectQuery = `SELECT short_url, original_url, deleted_flag FROM urls WHERE short_url = $1 or original_url = $1`

// SaveBatchInsertQuery insert query for batch save urls
const SaveBatchInsertQuery = `INSERT INTO urls (short_url, original_url, user_id) 
			 	VALUES %s
				ON CONFLICT (original_url) DO UPDATE SET short_url = EXCLUDED.short_url, original_url = EXCLUDED.original_url
				RETURNING short_url`

// DeleteBatchQuery delete urls by user
const DeleteBatchQuery = `UPDATE urls
			SET deleted_flag = true
			WHERE user_id = $1 and short_url = ANY($2)`

// GetAllUrlsByUserSelectQuery get all urls by user
const GetAllUrlsByUserSelectQuery = `SELECT short_url, original_url FROM urls WHERE user_id = $1 and length(short_url) > 0`

// NewDB factory for create DB storage
func NewDB(DBPath string) *DBStorage {
	db, err := sql.Open("pgx", DBPath)
	if err != nil {
		log.Fatal(err)
	}

	createUrlsTableQuery := `CREATE TABLE IF NOT EXISTS urls(
		id serial primary key,
		user_id uuid not null,
		short_url varchar(255) not null,
		original_url varchar(255) not null);
                               
	    CREATE UNIQUE INDEX IF NOT EXISTS original_url_idx ON urls (original_url);`
	_, err = db.Exec(createUrlsTableQuery)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	addColumnDeletedFlag := `ALTER TABLE urls ADD COLUMN IF NOT EXISTS deleted_flag boolean`
	_, err = db.Exec(addColumnDeletedFlag)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return &DBStorage{
		db: db,
	}
}

// Save function for save URL in DB
func (d *DBStorage) Save(ctx context.Context, originalURL string, shortURL string, userID *uuid.UUID) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var insertedURL string
	var pgErr *pgconn.PgError

	//	insertQuery := `INSERT INTO urls(short_url, original_url, user_id)
	//VALUES ($1, $2, $3)
	//RETURNING short_url`
	err := d.db.QueryRowContext(ctx, SaveInsertQuery, shortURL, originalURL, userID).Scan(&insertedURL)
	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			err = ErrConflict
			return "", NewURLConflictError(shortURL, err)
		}
		return "", err
	}
	return insertedURL, nil
}

// Get function for get URL from DB
func (d *DBStorage) Get(inputURL string) (string, error) {
	var short, original string
	var deletedFlag sql.NullBool

	row := d.db.QueryRowContext(context.Background(), GetSelectQuery, inputURL)
	if err := row.Scan(&short, &original, &deletedFlag); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("cannot scan row: %w", err)
	}

	if deletedFlag.Bool {
		return "", NewAlreadyDeletedError(inputURL)
	}

	if inputURL == short {
		return original, nil
	}

	if inputURL == original {
		return short, nil
	}

	return "", nil
}

// SaveBatch function for saving URL list
func (d *DBStorage) SaveBatch(ctx context.Context, urls []models.StorageURL, userID *uuid.UUID) ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

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
		insertValues += fmt.Sprintf("($%d, $%d, '%v')", paramNumber+1, paramNumber+2, userID)

		args = append(args, url.ShortURL)
		args = append(args, url.OriginalURL)
		paramNumber += 2
	}

	query := fmt.Sprintf(SaveBatchInsertQuery, insertValues)

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

// DeleteBatch function for delete URLs list
func (d *DBStorage) DeleteBatch(ctx context.Context, userID *uuid.UUID, urls []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	urlList := new(pgtype.VarcharArray)

	if err = urlList.Set(urls); err != nil {
		return fmt.Errorf("ошибка при формировании списка url для удаления: %v", err)
	}

	res, err := d.db.ExecContext(ctx, DeleteBatchQuery, userID, urlList)
	log.Print(res)

	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

// GetAllUrlsByUser function for get all user's URLs
func (d *DBStorage) GetAllUrlsByUser(ctx context.Context, userID *uuid.UUID) ([]models.StorageURL, error) {
	storageURLs := make([]models.StorageURL, 0)
	rows, err := d.db.QueryContext(ctx, GetAllUrlsByUserSelectQuery, userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var store models.StorageURL
		err = rows.Scan(&store.ShortURL, &store.OriginalURL)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}
			return nil, nil
		}
		storageURLs = append(storageURLs, store)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return storageURLs, nil
}

// Ping function for ping DB connection
func (d *DBStorage) Ping(ctx context.Context) error {
	if err := d.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}
