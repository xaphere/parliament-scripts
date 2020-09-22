package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func main() {
	// download files with voting data from parliament.bg

	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.WarnLevel)
	//log.Formatter = &logrus.JSONFormatter{}

	storage := "parlament-collector/data/"
	votesFile := "parlament-collector/data/votes.json"

	// construct vote.json file
	//baseURL, err := url.Parse("https://www.parliament.bg")
	//if err != nil {
	//	log.WithError(err).Fatal("failed to construct base url")
	//}
	//createVoteFiles(ctx, baseURL, votesFile, log)

	// download original xls files
	//downloadVoteFiles(ctx, votesFile, storage, log)

	fileServerAddress, err := url.Parse("http://127.0.0.1:9000")
	if err != nil {
		log.WithError(err).Fatal("failed to construct file server url")
	}

	getter := newFileCache(fileServerAddress, &xlsTransformer{BaseURL: "http://127.0.0.1:8080/transform"})
	collector := &voteDataCollector{
		getter: getter,
		log:    log,
	}
	collectedData := collector.CollectData(ctx, votesFile)
	storeCollectedDataAsCSV(filepath.Join(storage, "statistics.csv"), collectedData, log)
	storeCollectedDataAsJSON(filepath.Join(storage, "statistics.json"), collectedData, log)
}

func storeCollectedDataAsCSV(saveFileLoc string, voteData []collectedData, log *logrus.Logger) {

	dataFile, err := os.Create(saveFileLoc)
	if err != nil {
		log.WithError(err).Fatal("failed to create save csv file")
	}
	defer dataFile.Close()

	csvWriter := csv.NewWriter(dataFile)

	csvWriter.Write([]string{
		"name",
		"party",
		string(Here),
		string(Registered),
		string(Absent),
		string(For),
		string(Against),
		string(Abstain),
		string(NoVote),
	})
	for _, data := range voteData {
		csvWriter.Write([]string{
			data.Name,
			data.Party,
			strconv.Itoa(data.Votes[Here]),
			strconv.Itoa(data.Votes[Registered]),
			strconv.Itoa(data.Votes[Absent]),
			strconv.Itoa(data.Votes[For]),
			strconv.Itoa(data.Votes[Against]),
			strconv.Itoa(data.Votes[Abstain]),
			strconv.Itoa(data.Votes[NoVote]),
		})
	}
	csvWriter.Flush()
}

func storeCollectedDataAsJSON(saveFileLoc string, voteData []collectedData, log *logrus.Logger) {
	f, err := os.Create(saveFileLoc)
	if err != nil {
		log.WithError(err).Fatal("failed to create json save file")
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(voteData)
	if err != nil {
		log.WithError(err).Fatal("failed to marshal json  file")
	}
}

type fileGetter interface {
	GetFile(ctx context.Context, fileURL string) (*csv.Reader, error)
}

type voteDataCollector struct {
	getter fileGetter
	log    *logrus.Logger
}

type collectedData struct {
	Name  string           `json:"name"`
	Party string           `json:"party"`
	Votes map[VoteType]int `json:"votes"`
}

func (col *voteDataCollector) CollectData(ctx context.Context, voteFileLoc string) []collectedData {

	logEntry := col.log.WithField("vote-file", voteFileLoc)
	voteLoc, err := readVotesFile(voteFileLoc)
	if err != nil {
		logEntry.WithError(err).Fatal("failed to read vote file")
	}

	votesPerPerson := map[string]struct {
		Party string
		Vote  map[VoteType]int
	}{}
	for _, loc := range voteLoc {
		logEntry := logEntry.WithFields(map[string]interface{}{
			"session": loc.SessionURL,
			"member":  loc.MemberVoteURL,
			"party":   loc.PartyVoteURL,
		})
		logEntry.Info("extracting voting data")
		if loc.MemberVoteURL == "" {
			continue
		}
		membCSV, err := col.getter.GetFile(ctx, loc.MemberVoteURL)
		if err != nil {
			logEntry.WithError(err).Error("failed to get member csv data")
			continue
		}

		individualVotes, err := extractIndividualVoteDataFromCSV(membCSV)
		if err != nil {
			logEntry.WithError(err).Error("failed to extract individual data")
			continue
		}

		for _, vote := range individualVotes {
			for _, vt := range vote.Votes {
				if _, ok := votesPerPerson[vote.Name]; !ok {
					votesPerPerson[vote.Name] = struct {
						Party string
						Vote  map[VoteType]int
					}{Party: vote.Party, Vote: map[VoteType]int{}}
				}
				votesPerPerson[vote.Name].Vote[vt] = votesPerPerson[vote.Name].Vote[vt] + 1
			}
		}
	}

	data := []collectedData{}
	for name, votes := range votesPerPerson {
		data = append(data, collectedData{
			Name:  name,
			Party: votes.Party,
			Votes: votes.Vote,
		})
	}
	return data
}

type fileCache struct {
	LocalBaseURL *url.URL
	RemoteGetter fileGetter
}

func newFileCache(localBaseURL *url.URL, remoteGetter fileGetter) *fileCache {
	return &fileCache{LocalBaseURL: localBaseURL, RemoteGetter: remoteGetter}
}

func (fc *fileCache) GetFile(ctx context.Context, fileURL string) (*csv.Reader, error) {
	fileName := getFileNameFromURL(fileURL)

	fURL, err := fc.LocalBaseURL.Parse(fileName)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Get(fURL.String())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusOK {
		fURL.Host = "storage:4000"
		fileURL = fURL.String()
	}

	return fc.RemoteGetter.GetFile(ctx, fileURL)
}

type xlsTransformer struct {
	BaseURL string
}

func (x *xlsTransformer) GetFile(ctx context.Context, fileURL string) (*csv.Reader, error) {
	reqBody := fmt.Sprintf(`{"fileURL":"%s"}`, fileURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, x.BaseURL, strings.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to extract csv data")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	return csv.NewReader(bytes.NewReader(data)), nil
}
