package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/xaphere/parlament-scripts/pkg/parliament/models"
)

type SQLStorage struct {
	db *pgx.Conn
}

func NewDBConnection(baseURL string) (*SQLStorage, error) {
	db, err := pgx.Connect(context.Background(), baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to establish db connection: %w", err)
	}
	return &SQLStorage{
		db: db,
	}, nil
}

// CRUD Member
const membersTable = "members"

func (s *SQLStorage) CreateMember(ctx context.Context, member models.Member) error {
	query, params, err := insertQuery(membersTable, map[string]interface{}{
		"id":           member.ID,
		"name":         member.Name,
		"party":        member.PartyID,
		"constituency": member.ConstituencyID,
		"email":        member.Email,
	})
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, query, params...)
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLStorage) ReadMember(ctx context.Context, memberID int) (*models.Member, error) {
	var (
		id             int
		name           string
		partyID        int
		constituencyID int
		email          sql.NullString
	)
	err := s.db.QueryRow(ctx, "SELECT * FROM $1 WHERE id = $2", membersTable, memberID).
		Scan(&id, &name, &partyID, &constituencyID, &email)
	if err != nil {
		return nil, err
	}
	return &models.Member{
		ID:             id,
		Name:           name,
		PartyID:        partyID,
		ConstituencyID: constituencyID,
		Email:          email.String,
	}, nil

}

func (s *SQLStorage) UpdateMember(ctx context.Context, member models.Member) error {
	return errors.New("not implemented")
}

func (s *SQLStorage) DeleteMember(ctx context.Context, memberID int) error {
	return errors.New("not implemented")
}

// CRUD Party
const partyTable = "parliamentary_group"

func (s *SQLStorage) CreateParty(ctx context.Context, party models.Party) error {
	query, params, err := insertQuery(partyTable, map[string]interface{}{
		"id":   party.ID,
		"name": party.Name,
	})
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, query, params...)
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLStorage) ReaParty(ctx context.Context, partyID int) (*models.Party, error) {
	var (
		id   int
		name string
	)
	err := s.db.QueryRow(ctx, "SELECT * FROM $1 WHERE id = $2", partyTable, partyID).
		Scan(&id, &name)
	if err != nil {
		return nil, err
	}
	return &models.Party{
		ID:   id,
		Name: name,
	}, nil
}

func (s *SQLStorage) UpdateParty(ctx context.Context, party models.Party) error {
	return errors.New("not implemented")
}

func (s *SQLStorage) DeleteParty(ctx context.Context, partyID int) error {
	return errors.New("not implemented")
}

// CRUD constituency
const constituencyTable = "constituency"

func (s *SQLStorage) CreateConstituency(ctx context.Context, constituency models.Constituency) error {
	query, params, err := insertQuery(constituencyTable, map[string]interface{}{
		"id":   constituency.ID,
		"name": constituency.Name,
	})
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, query, params...)
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLStorage) ReaConstituency(ctx context.Context, constituencyID int) (*models.Constituency, error) {
	var (
		id   int
		name string
	)
	err := s.db.QueryRow(ctx, "SELECT * FROM $1 WHERE id = $2", constituencyTable, constituencyID).
		Scan(&id, &name)
	if err != nil {
		return nil, err
	}
	return &models.Constituency{
		ID:   id,
		Name: name,
	}, nil
}

func (s *SQLStorage) UpdateConstituency(ctx context.Context, constituency models.Constituency) error {
	return errors.New("not implemented")
}

func (s *SQLStorage) DeleteConstituency(ctx context.Context, constituencyID int) error {
	return errors.New("not implemented")
}

func (s *SQLStorage) CreateProceeding(ctx context.Context, proceeding *models.Proceeding) error {

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = txCreateProceeding(ctx, tx, proceeding)
	if err != nil {
		return err
	}
	_ = tx.Commit(ctx)
	return nil
}

func (s *SQLStorage) ReadProceeding(ctx context.Context, proceedingID models.ProceedingID) (*models.Proceeding, error) {

	var (
		id          string
		name        string
		date        time.Time
		locURL      string
		transcript  string
		attachments []string
		programID   sql.NullString
	)
	err := s.db.QueryRow(ctx, "SELECT * FROM proceedings WHERE id = $1", proceedingID).
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
		Program:     nil,
		Votes:       nil,
	}, nil
}

func txCreateProceeding(ctx context.Context, tx pgx.Tx, proceeding *models.Proceeding) error {
	attachments := []string{}
	for _, u := range proceeding.Attachments {
		attachments = append(attachments, u.String())
	}
	query, params, err := insertQuery("proceedings", map[string]interface{}{
		"id":          string(proceeding.UID),
		"name":        proceeding.Name,
		"date":        proceeding.Date,
		"url":         proceeding.URL.String(),
		"transcript":  proceeding.Transcript,
		"attachments": attachments,
	})
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, query, params...)
	if err != nil {
		return err
	}
	return nil
}

func insertQuery(table string, fields map[string]interface{}) (string, []interface{}, error) {
	keys := []string{}
	params := []interface{}{}
	placeholers := []string{}
	itr := 1
	for key, val := range fields {
		params = append(params, val)
		keys = append(keys, key)
		placeholers = append(placeholers, fmt.Sprintf("$%d", itr))
		itr++
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(keys, ", "), strings.Join(placeholers, ", ")), params, nil
}
