package store

import (
	"context"
	"database/sql"
)

type DBStore struct {
	*sql.DB
}

func NewDBStore(dsn string) (*DBStore, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	dbStore := &DBStore{DB: db}

	if err := dbStore.CreateTable(); err != nil {
		return nil, err
	}

	return dbStore, nil
}

func (db *DBStore) Get(id string) (string, error) {
	row := db.QueryRowContext(context.Background(), "SELECT url FROM shortener WHERE id = $1", id)
	var result string
	err := row.Scan(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (db *DBStore) Put(id string, url string) error {
	_, err := db.ExecContext(context.Background(), "INSERT INTO shortener VALUES ($1, $2)", id, url)
	return err
}

func (db *DBStore) DropTable() error {
	_, err := db.ExecContext(context.Background(), "DROP TABLE IF EXISTS shortener;")
	return err
}

func (db *DBStore) CreateTable() error {
	_, err := db.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS shortener( "+
		"id VARCHAR(255) PRIMARY KEY, "+
		"url VARCHAR(255) NOT NULL "+
		");")
	return err
}
