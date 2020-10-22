package models

type MemberID string
type Member struct {
	UID   MemberID
	Name  string
	Party string
}
