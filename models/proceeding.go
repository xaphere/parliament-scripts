package models

import (
	"net/url"
	"time"
)

type ProceedingID string

type Proceeding struct {
	// unique proceeding id given from the parliament system
	// example: 10474
	UID ProceedingID
	// name of the proceeding
	// example: "ЧЕТИРИСТОТИН И ТРЕТО ЗАСЕДАНИЕ"
	Name string
	// Date and time of the proceeding
	// example 2020-09-23T13:03:00Z03:00
	Date time.Time
	// online location of the proceeding data
	// example: "https://www.parliament.bg/bg/plenaryst/ns/52/ID/10474"
	URL *url.URL
	// full text transcript of the proceeding, as provided by the parliament system
	Transcript string
	// locations of all attached files for the proceeding
	Attachments []*url.URL
	// id of the planed assembly program as assigned by the parliament system
	// example: 1135 for https://www.parliament.bg/bg/plenaryprogram/ID/1135
	ProgID ProgramID
	// collection of result for every vote that took place in the proceeding
	Votes []struct {
		// unique id of the vote constructed of the proceeding ID and vote number
		// example: "10474-3"
		UID VoteID
		//
		Results map[MemberID]VoteType
	}
}
