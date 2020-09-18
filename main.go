package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

const extractorLocation = "http://127.0.0.1:8080/transform"
const baseParliamentURL = "https://www.parliament.bg"
const plenaryStURL = "/bg/plenaryst"

func main() {

}

func getAllVoteURLs() {
	log := logrus.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log.Info("Start scraping")
	pageData, err := getPage(ctx, baseParliamentURL+plenaryStURL)
	if err != nil {
		log.WithError(err).Fatal("failed to get base plenary page")
	}

	periods := getSessionPeriods(pageData)

	allSessions := map[string]struct{}{}
	count := 0
	for _, period := range periods {
		log.WithField("url", baseParliamentURL+period).Info("Request Page")
		pageData, err := getPage(ctx, baseParliamentURL+period)
		if err != nil {
			log.WithError(err).WithField("period", period).Error("failed to get period page")
			continue
		}
		sessions := getSessions(pageData)
		for _, sess := range sessions {
			allSessions[sess] = struct{}{}
			count++
		}
		time.Sleep(time.Second)
	}
	log.Infof("number of sessions: %d", count)
	type VotingURLs struct {
		SessionURL    string `json:"session_url"`
		MemberVoteURL string `json:"member_vote_url"`
		PartyVoteURL  string `json:"party_vote_url"`
	}

	votings := []VotingURLs{}
	failedSessions := []string{}
	for sess := range allSessions {
		log.WithField("url", baseParliamentURL+sess).Info("Request Page")
		pageData, err := getPage(ctx, baseParliamentURL+sess)
		if err != nil {
			log.WithError(err).WithField("session", sess).Error("failed to get sessions page")
			failedSessions = append(failedSessions, sess)
			continue
		}
		memVote, partyVote := getVotingURLs(pageData)
		votings = append(votings, VotingURLs{
			SessionURL:    baseParliamentURL + sess,
			MemberVoteURL: baseParliamentURL + memVote,
			PartyVoteURL:  baseParliamentURL + partyVote,
		})
		time.Sleep(time.Second * 5)
	}

	f, err := os.Create("votes.json")
	if err != nil {
		log.WithError(err).Fatal("failed to create file")
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(votings)
	if err != nil {
		log.WithError(err).Fatal("failed to marshal voting data")
	}
}
