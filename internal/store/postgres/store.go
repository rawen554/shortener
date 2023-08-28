package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rawen554/shortener/internal/models"
)

// DBStore - Интерфейс работы с пулом соединений.
type DBStore struct {
	conn *pgxpool.Pool
}

// CPUMultiplyer Мультипликатор для конфигурации максимального кол-ва соединений.
const CPUMultiplyer = 4

// ErrDBInsertConflict Обнаружен конфликт в БД, необходимо его обработать.
var ErrDBInsertConflict = errors.New("conflict insert into table, returned stored value")

// ErrURLDeleted Запрашиваемый URL удален.
var ErrURLDeleted = errors.New("url is deleted")

// NewPostgresStore Функция получения экземпляра DBStore.
func NewPostgresStore(ctx context.Context, dsn string) (*DBStore, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}

	conf, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config from string: %w", err)
	}

	conf.MaxConns = int32(runtime.NumCPU() * CPUMultiplyer)

	conn, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create new conn pool: %w", err)
	}
	dbStore := &DBStore{conn: conn}

	return dbStore, nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

// Применение миграций из папки в текущем каталоге - migrations.
func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

func (db *DBStore) Ping() error {
	if err := db.conn.Ping(context.Background()); err != nil {
		return fmt.Errorf("lost connection to db: %w", err)
	}
	return nil
}

func (db *DBStore) Close() {
	db.conn.Close()
}

func (db *DBStore) Get(id string) (string, error) {
	row := db.conn.QueryRow(context.Background(), "SELECT original_url, deleted_flag FROM shortener WHERE slug = $1", id)
	var result string
	var deleted bool
	err := row.Scan(&result, &deleted)
	if err != nil {
		return "", fmt.Errorf("cant scan result: %w", err)
	}

	if deleted {
		return "", ErrURLDeleted
	}

	return result, nil
}

func (db *DBStore) GetAllByUserID(userID string) ([]models.URLRecord, error) {
	result := make([]models.URLRecord, 0)

	rows, err := db.conn.Query(context.Background(), `
		SELECT slug, original_url
		FROM shortener
		WHERE user_id = $1 AND deleted_flag = FALSE
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query all users records: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		record := models.URLRecord{}
		if err := rows.Scan(&record.ShortURL, &record.OriginalURL); err != nil {
			return nil, fmt.Errorf("cant scan records: %w", err)
		}

		result = append(result, record)
	}

	return result, nil
}

func (db *DBStore) DeleteMany(ids models.DeleteUserURLsReq, userID string) error {
	ctx := context.Background()

	query := `
		UPDATE shortener SET deleted_flag = TRUE
		WHERE shortener.slug = $1 AND shortener.user_id = $2`
	batch := &pgx.Batch{}
	for _, url := range ids {
		batch.Queue(query, url, userID)
	}
	batchResults := db.conn.SendBatch(ctx, batch)
	defer func() {
		if err := batchResults.Close(); err != nil {
			log.Printf("error closing result: %v", err)
		}
	}()

	for range ids {
		_, err := batchResults.Exec()
		if err != nil {
			log.Printf("error executing: %v", err)
			return fmt.Errorf("cant exec batch: %w", err)
		}
	}

	return nil
}

func (db *DBStore) Put(id string, url string, userID string) (string, error) {
	var err error

	row := db.conn.QueryRow(context.Background(), `
		INSERT INTO shortener VALUES ($1, $2, $3)
		ON CONFLICT (original_url)
		DO UPDATE SET
			original_url=EXCLUDED.original_url
		RETURNING slug
	`, id, url, userID)
	var result string
	if err := row.Scan(&result); err != nil {
		return "", fmt.Errorf("cant scan put record result: %w", err)
	}

	if id != result {
		err = ErrDBInsertConflict
	}

	return result, err
}

func (db *DBStore) PutBatch(urls []models.URLBatchReq, userID string) ([]models.URLBatchRes, error) {
	query := `
		INSERT INTO shortener VALUES (@slug, @originalUrl, @userID)
		ON CONFLICT (original_url)
		DO UPDATE SET
			original_url=EXCLUDED.original_url
		RETURNING slug
	`
	result := make([]models.URLBatchRes, 0)

	batch := &pgx.Batch{}
	for _, url := range urls {
		args := pgx.NamedArgs{
			"slug":        url.CorrelationID,
			"originalUrl": url.OriginalURL,
			"userID":      userID,
		}
		batch.Queue(query, args)
	}
	results := db.conn.SendBatch(context.Background(), batch)
	defer func() {
		if err := results.Close(); err != nil {
			log.Printf("error closing batch result: %v", err)
		}
	}()

	for _, url := range urls {
		id, err := results.Exec()
		if err != nil {
			return nil, fmt.Errorf("cant exec tx: %w", err)
		}
		result = append(result, models.URLBatchRes{
			CorrelationID: url.CorrelationID,
			ShortURL:      id.String(),
		})
	}

	return result, nil
}
