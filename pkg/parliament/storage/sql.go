package storage

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v4"
)

type SQLStorage struct {
	dbMX *sync.RWMutex
	db   *sql.DB

	BaseURL string
}

func NewDB(baseURL string) *SQLStorage {
	return &SQLStorage{
		dbMX:    &sync.RWMutex{},
		BaseURL: baseURL,
	}
}

func (s *SQLStorage) Connect(ctx context.Context) error {
	db, err := pgx.Connect(ctx, s.BaseURL)
	if err != nil {
		return fmt.Errorf("failed to establish db connection: %w", err)
	}
	s.setDB(db)
	return nil
}

func (s *SQLStorage) Disconnect(ctx context.Context) {
	db := s.db
	s.setDB(nil)
	db.Close(ctx)
}

func (s *SQLStorage) setDB(db *sql.DB) {
	s.dbMX.Lock()
	defer s.dbMX.Unlock()
	s.db = db
}
