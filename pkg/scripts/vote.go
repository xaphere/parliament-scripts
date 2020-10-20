package scripts

import (
	"encoding/csv"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type VoteID int

// Vote represents a single record for what the MPs voted in a particular session
type Vote struct {
	ID    VoteID    // unique per session id
	Date  time.Time // time and date when the vote took plase
	Title string    // What was voted for
}

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

var voteMapTable = map[string]VoteType{
	"+": For,
	"=": Abstain,
	"-": Against,
	"0": NoVote,
	"П": Here,
	"О": Absent,
	"Р": Registered,
}

// Individual represents how a particular parliament member voted for on every issue in a session
type Individual struct {
	Number int
	Name   string
	Party  string
	Votes  map[VoteID]VoteType
}

var ErrVoteNotFound = errors.New("no matches found")

func constructVoteDataFormString(data string) (*Vote, error) {
	re := regexp.MustCompile(`Номер \((?P<id>\d+)\) (?P<type>\p{L}+) проведено на (?P<date>[\d\s:-]+) по тема (?P<title>.*)`)
	const template = `$id|$type|$date|$title`
	result := []byte{}
	submatch := re.FindAllStringSubmatchIndex(data, -1)
	if len(submatch) != 1 {
		return nil, ErrVoteNotFound
	}
	extracted := re.ExpandString(result, template, data, submatch[0])
	str := strings.Split(string(extracted), "|")
	if len(str) != 4 {
		return nil, errors.New("failed to extract valid data")
	}
	id, err := strconv.Atoi(str[0])
	if err != nil {
		return nil, err
	}
	date, err := time.Parse(`02-01-2006 15:04`, str[2])
	if err != nil {
		return nil, err
	}
	return &Vote{
		ID:    VoteID(id),
		Date:  date,
		Title: str[3],
	}, nil
}

func ExtractVoteDataFromCSV(reader *csv.Reader) ([]Vote, error) {
	const voteColumn = 1
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	result := []Vote{}
	for _, roll := range data {
		if len(roll) <= voteColumn+1 {
			return nil, errors.New("invalid csv format")
		}
		c := roll[voteColumn]
		voteData, err := constructVoteDataFormString(c)
		if err != nil {
			if errors.Is(err, ErrVoteNotFound) {
				continue
			}
			fmt.Printf("failed to extract voteData: %s : %v\n", c, err)
			continue
		}
		result = append(result, *voteData)
	}
	return result, nil
}

func ExtractIndividualVoteDataFromCSV(reader *csv.Reader) ([]Individual, error) {
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}
	voteColRelation := map[int]VoteID{}
	firstVoteCol := -1
	for idx, h := range headers {
		id, err := strconv.Atoi(h)
		if err != nil {
			continue
		}
		if id == 0 {
			// we start the count from 1
			continue
		}
		if id == 1 {
			firstVoteCol = idx
		}
		voteColRelation[idx] = VoteID(id)
	}

	// example first vote is the registration one 'П'
	// 1,АДЛЕН ШУКРИ ШЕВКЕД,,1245.0,ДПС,П,П,+
	//
	// member name is at 1
	// the party collum is at firstVoteCol-1
	// member number is at firstVoteCol-2
	// So we need at least 4 columns to extract any data
	if firstVoteCol <= 3 {
		return nil, errors.New("insufficient data")
	}
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	result := []Individual{}
	var nameCol = 1
	var partyCol = firstVoteCol - 1
	var numberCol = firstVoteCol - 2
	for _, roll := range records {
		name := roll[nameCol]
		party := roll[partyCol]
		n := roll[numberCol]
		num, err := strconv.ParseFloat(n, 0)
		if err != nil {
			fmt.Printf("failed extract number for %s\n", name)
		}

		member := Individual{
			Number: int(num),
			Name:   name,
			Party:  party,
			Votes:  map[VoteID]VoteType{},
		}
		for colIdx, voteID := range voteColRelation {
			vote, ok := voteMapTable[roll[colIdx]]
			if !ok {
				vote = VoteType("unknown type: " + roll[colIdx])
			}
			member.Votes[voteID] = vote
		}

		result = append(result, member)
	}
	return result, nil
}
