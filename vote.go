package main

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

var ErrVoteNotFound = errors.New("no matches found")

func extractVoteDataFormString(data string) (*Vote, error) {
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

func extractVoteDataFromCSV(reader *csv.Reader) ([]Vote, error) {
	const voteCollum = 1
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	result := []Vote{}
	for _, roll := range data {
		if len(roll) <= voteCollum+1 {
			return nil, errors.New("invalid csv format")
		}
		c := roll[voteCollum]
		voteData, err := extractVoteDataFormString(c)
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
