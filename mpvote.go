package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
)

type VoteType string

const For VoteType = "for"
const Against VoteType = "against"
const Abstain VoteType = "abstain"
const Missing VoteType = "missing"

const Here VoteType = "here"
const Absent VoteType = "absent"

// MPRecord represents how a particular parlament member voted for on every issue in a session
type MPRecord struct {
	Number int
	Name   string
	Party  string
	Votes  map[VoteID]VoteType
}

func extractMPVoteDataFromCSV(reader *csv.Reader) ([]MPRecord, error) {
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
	result := []MPRecord{}
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

		member := MPRecord{
			Number: int(num),
			Name:   name,
			Party:  party,
			Votes:  map[VoteID]VoteType{},
		}
		for colIdx, voteID := range voteColRelation {
			vote := Missing
			switch roll[colIdx] {
			case "+":
				vote = For
			case "=":
				vote = Abstain
			case "-":
				vote = Against
			case "0":
				vote = Missing
			case "П":
				vote = Here
			case "О":
				vote = Absent
			default:
				vote = VoteType("unknown type: " + roll[colIdx])
			}
			member.Votes[voteID] = vote
		}

		result = append(result, member)
	}
	return result, nil
}
