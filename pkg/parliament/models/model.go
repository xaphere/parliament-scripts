package models

import (
	"net/url"
	"time"
)

type VoteType string

const For VoteType = "for"
const Against VoteType = "against"
const Abstain VoteType = "abstain"
const NoVote VoteType = "no-vote"

// Here, Registered and Absent are used when the assembly counts the MPs to see if there is quorum
// Here and Absent are self-explanatory,
// But I'm not sure what "Р" means. By cross referencing with the per party vote table,
// I deduced that "Р" is counted as absent for the purposes of a quorum.
// I think is when the MP has put their card in the voting terminal but did not press a button when called for.
const Here VoteType = "here"
const Registered VoteType = "registered"
const Absent VoteType = "absent"

type VoteID string

// Vote represents a single record for what the MPs voted in a proceeding
type Vote struct {
	UID   VoteID
	Date  time.Time // time and date when the vote took place
	Title string    // What was voted for
}

type MemberID string

// Member represents a parliament member data
type Member struct {
	UID   MemberID
	Name  string
	Party string
}

// example: https://www.parliament.bg/bg/plenaryprogram/ID/1135
type ProgramID string

// Proceedings program is the planed program for the assembly to vote on
type ProceedingProgram struct {
	UID    ProgramID
	Form   time.Time
	To     time.Time
	Points struct {
		Name        string
		Attachments []url.URL
	}
}

type ProceedingID string

// Proceeding is all the public data for a assembly proceeding
type Proceeding struct {
	// unique proceeding id given from the parliament system
	// example: 10474
	UID ProceedingID `db`
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
