package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const extractorLocation = "http://127.0.0.1:8080/transform"
const baseParliamentURL = "https://www.parliament.bg"
const plenaryStURL = "/bg/plenaryst"

func main() {
	// download files with voting data from parliament.bg

	ctx := context.Background()
	log := logrus.New()
	log.Formatter = &logrus.JSONFormatter{}
	//
	//storage := "parlament-collector/data/originals"
	votesFile := "parlament-collector/data/votes.json"
	//downloadVoteFiles(ctx, votesFile, storage, log)

	baseURL, err := url.Parse("https://www.parliament.bg")
	if err != nil {
		log.WithError(err).Fatal("failed to construct base url")
	}
	createVoteFiles(ctx, baseURL, votesFile, log)
}

func extractVoteData() {
	const dataLoc = "parlament-collector/data/"
	log := logrus.New()
	f, err := os.Open(dataLoc + "vote_data.csv")
	if err != nil {
		log.WithError(err).Fatal("failed to open vote_data.csv")
	}
	defer f.Close()
	reader := csv.NewReader(f)

	for {
		rec, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Info("finished reading the file")
				break
			} else if errors.Is(err, csv.ErrFieldCount) {
				log.WithError(err).WithField("count", len(rec)).Warn()
			} else {
				log.WithError(err).Error("unexpected error")
				break
			}
		}
		printVoteData(rec, log)

	}
}

func printVoteData(rec []string, log *logrus.Logger) {
	Name := rec[0]
	Votes := map[string]int{}
	for idx := 1; idx < len(rec); idx++ {
		vote := rec[idx]
		Votes[vote] = Votes[vote] + 1
	}
	persent := map[string]interface{}{}
	maxVote := len(rec) - 1
	for v, count := range Votes {
		persent[v] = (float32(count) / float32(maxVote)) * 100
	}

	log.WithField("name", Name).WithFields(persent).Info()
}

func collectVotingData() {
	const dataLoc = "parlament-collector/data/"
	log := logrus.New()

	var voteLoc []VotingURLs
	voteFile, err := os.Open(dataLoc + "votes.json")
	if err != nil {
		log.WithError(err).Fatal("failed to open  data/votes.json")
	}
	err = json.NewDecoder(voteFile).Decode(&voteLoc)
	if err != nil {
		log.WithError(err).Fatal("failed to unmarshal vote url data")
	}
	voteFile.Close()

	col, err := NewCollector(extractorLocation, baseParliamentURL)
	if err != nil {
		log.WithError(err).Fatal("failed to construct collector")
	}
	ctx := context.WithValue(context.Background(), deltaKey, 5)

	votes := map[string][]string{}
	for _, loc := range voteLoc {
		log.WithField("url", loc.SessionURL).Info("extracting voting data")
		data, err := col.GetVotingData(ctx, loc.SessionURL)
		if err != nil {
			log.WithError(err).WithField("url", loc.SessionURL).Error("failed to extract voting data")
			continue
		}

		for _, mp := range data.MPData {
			for _, vote := range mp.Votes {
				votes[mp.Name] = append(votes[mp.Name], string(vote))
			}
		}

		slashIdx := strings.LastIndex(loc.SessionURL, "/")
		id := loc.SessionURL[slashIdx+1:]
		f, err := os.Create(dataLoc + "mpData_" + id + ".json")
		if err != nil {
			log.WithError(err).WithField("url", loc.SessionURL).Error("failed to create file to store voting data")
			continue
		}
		err = json.NewEncoder(f).Encode(data)
		f.Close()
		if err != nil {
			log.WithError(err).WithField("url", loc.SessionURL).Error("failed to marshal voting data to file")
		}
		time.Sleep(time.Second * 10)
	}

	voteDataFile, err := os.Create(dataLoc + "vote_data.csv")
	if err != nil {
		log.WithError(err).Fatal("failed to create vote data csv file")
	}
	defer voteDataFile.Close()
	csvWriter := csv.NewWriter(voteDataFile)
	for name, v := range votes {
		record := []string{name}
		record = append(record, v...)
		csvWriter.Write(record)
	}
	csvWriter.Flush()
}

type VotingURLs struct {
	SessionURL    string `json:"session_url"`
	MemberVoteURL string `json:"member_vote_url"`
	PartyVoteURL  string `json:"party_vote_url"`
}
