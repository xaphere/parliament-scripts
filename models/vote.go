package models

import "time"

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

// Vote represents a single record for what the MPs voted in a particular session
type Vote struct {
	UID   VoteID
	Date  time.Time // time and date when the vote took place
	Title string    // What was voted for
}
