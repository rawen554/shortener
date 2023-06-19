package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rawen554/shortener/internal/models"
)

type DBStore struct {
	conn *pgx.Conn
}

var ErrDBInsertConflict = errors.New("conflict insert into table, returned stored value")

func NewPostgresStore(dsn string) (*DBStore, error) {
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	dbStore := &DBStore{conn: conn}

	if err := dbStore.CreateTable(); err != nil {
		return nil, err
	}

	return dbStore, nil
}

func (db *DBStore) HealthCheck() error {
	return db.conn.Ping(context.Background())
}

func (db *DBStore) Get(id string) (string, error) {
	row := db.conn.QueryRow(context.Background(), "SELECT url FROM shortener WHERE id = $1", id)
	var result string
	err := row.Scan(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (db *DBStore) Put(id string, url string) (string, error) {
	var err error

	row := db.conn.QueryRow(context.Background(), `
		INSERT INTO shortener VALUES ($1, $2)
		ON CONFLICT (url)
		DO UPDATE SET
			url=EXCLUDED.url
		RETURNING id
	`, id, url)
	var result string
	if err := row.Scan(&result); err != nil {
		return "", err
	}

	if id != result {
		err = ErrDBInsertConflict
	}

	return result, err
}

func (db *DBStore) PutBatch(urls []models.URLBatchReq) ([]models.URLBatchRes, error) {
	query := `
		INSERT INTO shortener VALUES (@id, @originalUrl)
		ON CONFLICT (url)
		DO UPDATE SET
			url=EXCLUDED.url
		RETURNING id	
	`
	result := make([]models.URLBatchRes, 0)

	batch := &pgx.Batch{}
	for _, url := range urls {
		args := pgx.NamedArgs{
			"id":          url.CorrelationID,
			"originalUrl": url.OriginalURL,
		}
		batch.Queue(query, args)
	}
	results := db.conn.SendBatch(context.Background(), batch)
	defer results.Close()

	for _, url := range urls {
		id, err := results.Exec()
		if err != nil {
			return nil, err
		}
		result = append(result, models.URLBatchRes{
			CorrelationID: url.CorrelationID,
			ShortURL:      id.String(),
		})
	}

	return result, nil
}

func (db *DBStore) CreateTable() error {
	_, err := db.conn.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS shortener( "+
		"id VARCHAR(255) NOT NULL, "+
		"url VARCHAR(255) PRIMARY KEY "+
		");")
	return err
}
