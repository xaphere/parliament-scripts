package main

import (
	"bytes"
	"context"
	"encoding/csv"
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
	collector.CollectData(ctx, votesFile, filepath.Join(storage, "statistics.csv"))
}

type fileGetter interface {
	GetFile(ctx context.Context, fileURL string) (*csv.Reader, error)
}

type voteDataCollector struct {
	getter fileGetter
	log    *logrus.Logger
}

func (col *voteDataCollector) CollectData(ctx context.Context, voteFileLoc string, saveFileLoc string) {

	logEntry := col.log.WithField("vote-file", voteFileLoc).WithField("save-file", saveFileLoc)
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

	dataFile, err := os.Create(saveFileLoc)
	if err != nil {
		logEntry.WithError(err).Fatal("failed to create save file")
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
	for name, votes := range votesPerPerson {
		csvWriter.Write([]string{
			name,
			votes.Party,
			strconv.Itoa(votes.Vote[Here]),
			strconv.Itoa(votes.Vote[Registered]),
			strconv.Itoa(votes.Vote[Absent]),
			strconv.Itoa(votes.Vote[For]),
			strconv.Itoa(votes.Vote[Against]),
			strconv.Itoa(votes.Vote[Abstain]),
			strconv.Itoa(votes.Vote[NoVote]),
		})
	}
	csvWriter.Flush()
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
