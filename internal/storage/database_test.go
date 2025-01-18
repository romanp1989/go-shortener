package storage

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/pgtype"
	"github.com/romanp1989/go-shortener/internal/auth"
	"github.com/romanp1989/go-shortener/internal/models"
	"slices"
	"sync"
	"testing"
)

func TestDBStorage_Save(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := DBStorage{
		db: db,
		mu: sync.RWMutex{},
	}

	type args struct {
		ctx         context.Context
		originalURL string
		shortURL    string
		userID      *uuid.UUID
	}

	userID := auth.EnsureRandom()

	mock.ExpectQuery("INSERT INTO urls").
		WithArgs("6YGS4ZUF", "https://ya.ru", userID).
		WillReturnRows(sqlmock.NewRows([]string{"short_url"}).AddRow("6YGS4ZUF"))

	tests := []struct {
		name string
		//fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Success_Save",
			//fields: testFields,
			args: args{
				ctx:         context.Background(),
				originalURL: "https://ya.ru",
				shortURL:    "6YGS4ZUF",
				userID:      &userID,
			},
			want:    "6YGS4ZUF",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.Save(tt.args.ctx, tt.args.originalURL, tt.args.shortURL, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Save() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBStorage_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := DBStorage{
		db: db,
		mu: sync.RWMutex{},
	}

	type args struct {
		inputURL string
	}

	mock.ExpectQuery("SELECT short_url, original_url, deleted_flag FROM urls").
		WithArgs("6YGS4ZUF").
		WillReturnRows(sqlmock.NewRows([]string{"short_url", "original_url", "deleted_flag"}).AddRow("6YGS4ZUF", "https://ya.ru", false))
	mock.ExpectQuery("SELECT short_url, original_url, deleted_flag FROM urls").
		WithArgs("https://ya.ru").
		WillReturnRows(sqlmock.NewRows([]string{"short_url", "original_url", "deleted_flag"}).AddRow("6YGS4ZUF", "https://ya.ru", false))

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Success_Get_original_URL",
			args: args{
				inputURL: "6YGS4ZUF",
			},
			want:    "https://ya.ru",
			wantErr: false,
		},
		{
			name: "Success_Get_short_URL",
			args: args{
				inputURL: "https://ya.ru",
			},
			want:    "6YGS4ZUF",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.Get(tt.args.inputURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBStorage_SaveBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := DBStorage{
		db: db,
		mu: sync.RWMutex{},
	}

	type args struct {
		inputURLs []models.StorageURL
	}

	userID := auth.EnsureRandom()

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO urls").
		WithArgs("6YGS4ZUF", "https://ya.ru").
		WillReturnRows(sqlmock.NewRows([]string{"short_url"}).AddRow("6YGS4ZUF"))
	mock.ExpectCommit()

	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Success_SaveBatch",
			args: args{
				inputURLs: []models.StorageURL{
					{
						UserID:      &userID,
						OriginalURL: "https://ya.ru",
						ShortURL:    "6YGS4ZUF",
					},
				},
			},
			want:    []string{"https://ya.ru"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.SaveBatch(context.Background(), tt.args.inputURLs, &userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if slices.Equal(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBStorage_DeleteBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := DBStorage{
		db: db,
		mu: sync.RWMutex{},
	}

	type args struct {
		inputURLs []string
	}

	userID := auth.EnsureRandom()

	urlList := new(pgtype.VarcharArray)

	if err = urlList.Set([]string{"6YGS4ZUF"}); err != nil {
		t.Errorf("ошибка при формировании списка url для удаления: %v", err)
		return
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE urls").
		WithArgs(userID, urlList).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Success_DeleteBatch",
			args: args{
				inputURLs: []string{"6YGS4ZUF"},
			},
			want:    []string{"https://ya.ru"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.DeleteBatch(context.Background(), &userID, tt.args.inputURLs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
