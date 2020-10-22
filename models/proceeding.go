package models

import (
	"net/url"
	"time"
)

type ProceedingID string

type Proceeding struct {
	UID         ProceedingID
	Name        string
	Date        time.Time
	URL         url.URL
	Transcript  string
	Attachments []url.URL
	ProgramID   string
	Votes       []struct {
		UID     VoteID
		Results map[MemberID]VoteType
	}
}
