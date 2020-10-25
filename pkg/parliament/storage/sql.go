package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

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

func (s *SQLStorage) CreateProceeding(ctx context.Context, proceeding *models.Proceeding) error {
	conn, connClose := s.getDB()
	defer connClose()

	if conn == nil {
		return errors.New("no connection")
	}

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

func (s *SQLStorage) ReadProceeding(ctx context.Context, proceedingID models.ProceedingID) (*models.Proceeding, error) {

	conn, connClose := s.getDB()
	defer connClose()

	if conn == nil {
		return nil, errors.New("no connection")
	}

	var (
		id          string
		name        string
		date        time.Time
		locURL      string
		transcript  string
		attachments []string
		programID   sql.NullString
	)
	err := conn.QueryRow(ctx, "SELECT * FROM proceedings WHERE id = $1", proceedingID).
		Scan(&id, &name, &date, &locURL, &transcript, &attachments, &programID)
	if err != nil {
		return nil, err
	}

	//t, err := time.Parse(time.RFC3339, date)
	loc, err := url.Parse(locURL)

	att := []*url.URL{}
	for _, u := range attachments {
		a, err := url.Parse(u)
		if err != nil {
			continue
		}
		att = append(att, a)
	}

	return &models.Proceeding{
		UID:         models.ProceedingID(id),
		Name:        name,
		Date:        date,
		URL:         loc,
		Transcript:  transcript,
		Attachments: att,
		ProgID:      models.ProgramID(programID.String),
		Votes:       nil,
	}, nil
}
