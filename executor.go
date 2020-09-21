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
	"strings"
)

type VotingData struct {
	SourceURL         string
	IndividualVoteURL string
	PartyVoteURL      string

	Votes  []Vote
	MPData []MPRecord
}

type xlsTransformer interface {
	TransformFile(ctx context.Context, loc string) (*csv.Reader, error)
}
type Collector struct {
	transformer       xlsTransformer
	baseParliamentURL *url.URL
}

func NewCollector(transforerURL, baseURL string) (*Collector, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return &Collector{
		transformer:       &extractor{baseURL: transforerURL},
		baseParliamentURL: base,
	}, nil
}

func (c *Collector) GetVotingData(ctx context.Context, sourceURL string) (*VotingData, error) {

	pageData, err := getPage(ctx, sourceURL)
	if err != nil {
		return nil, err
	}

	ivURL, gvURL := getVotingURLs(pageData)
	if err != nil {
		return nil, err
	}

	memberURL, err := c.baseParliamentURL.Parse(ivURL)
	if err != nil {
		return nil, err
	}
	members, err := getMemberVoting(ctx, c.transformer, memberURL.String())
	if err != nil {
		return nil, err
	}

	partyURL, err := c.baseParliamentURL.Parse(gvURL)
	if err != nil {
		return nil, err
	}
	votes, err := getPartyVoteData(ctx, c.transformer, partyURL.String())
	if err != nil {
		return nil, err
	}

	return &VotingData{
		SourceURL:         sourceURL,
		IndividualVoteURL: memberURL.String(),
		PartyVoteURL:      partyURL.String(),
		Votes:             votes,
		MPData:            members,
	}, nil
}

func getMemberVoting(ctx context.Context, tr xlsTransformer, memberVotingURL string) ([]MPRecord, error) {
	reader, err := tr.TransformFile(ctx, memberVotingURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get vote by member file: %w", err)
	}
	members, err := extractMPVoteDataFromCSV(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to get votes: %w", err)
	}
	return members, nil
}

func getPartyVoteData(ctx context.Context, tr xlsTransformer, partyVotingURL string) ([]Vote, error) {
	reader, err := tr.TransformFile(ctx, partyVotingURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get vote by party file: %w", err)
	}
	votes, err := extractVoteDataFromCSV(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to extract votes: %w", err)
	}
	return votes, nil
}

type extractor struct {
	baseURL string
}

func (e *extractor) TransformFile(ctx context.Context, fileURL string) (*csv.Reader, error) {
	reqBody := fmt.Sprintf(`{"fileURL":"%s"}`, fileURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL, strings.NewReader(reqBody))
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
