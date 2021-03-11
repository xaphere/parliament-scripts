package models

// Member represents a parliament member data
type Member struct {
	ID             int
	Name           string
	PartyID        int
	ConstituencyID int
	Email          string
}

type Constituency struct {
	ID   int
	Name string
}

type Party struct {
	ID   int
	Name string
}
