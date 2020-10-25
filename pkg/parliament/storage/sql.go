package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v4"
	"github.com/xaphere/parlament-scripts/pkg/parliament/models"
)

type SQLStorage struct {
	dbMX *sync.RWMutex
	db   *pgx.Conn

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

func (s *SQLStorage) setDB(db *pgx.Conn) {
	s.dbMX.Lock()
	defer s.dbMX.Unlock()
	s.db = db
}

func (s *SQLStorage) getDB() (*pgx.Conn, func()) {
	s.dbMX.RLock()
	return s.db, s.dbMX.RUnlock
}

func (s *SQLStorage) StoreProceeding(ctx context.Context, proceeding *models.Proceeding) error {
	conn, connClose := s.getDB()
	defer connClose()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	attachment := []string{}
	for _, u := range proceeding.Attachments {
		attachment = append(attachment, u.String())
	}

	_, err = tx.Exec(ctx, `INSERT INTO proceedings (id, name, date, url, transcript, attachments) VALUES ($1, $2, $3, $4, $5, $6)`,
		proceeding.UID,
		proceeding.Name,
		proceeding.Date,
		proceeding.URL.String(),
		proceeding.Transcript,
		attachment,
	)
	if err != nil {
		return err
	}
	tx.Commit(ctx)
	return nil
}
