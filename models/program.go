package models

import (
	"net/url"
	"time"
)

// example: https://www.parliament.bg/bg/plenaryprogram/ID/1135
type ProgramID string
type ProceedingProgram struct {
	UID    ProgramID
	Form   time.Time
	To     time.Time
	Points struct {
		Name        string
		Attachments []url.URL
	}
}
