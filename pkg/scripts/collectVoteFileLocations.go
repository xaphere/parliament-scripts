package scripts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/sirupsen/logrus"
)

func collectAllSessionPeriods(ctx context.Context, parliamentURL *url.URL) ([]*url.URL, error) {
	sessionsBaseURL, err := parliamentURL.Parse("/bg/plenaryst")
	if err != nil {
		return nil, fmt.Errorf("can't construct transcript location: %w", err)
	}

	sessionsPageData, err := RequestPage(ctx, sessionsBaseURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get base transcript page: %w", err)

	}
	sessionsPage := string(sessionsPageData)
	sessionPeriods := getSessionPeriodURLs(sessionsPage)
	result := []*url.URL{}
	for _, loc := range sessionPeriods {
		pageURL, err := parliamentURL.Parse(loc)
		if err != nil {
			return nil, fmt.Errorf("failed to construct session period url: %w", err)
		}
		result = append(result, pageURL)
	}
	return result, nil
}

func collectSessionURLs(ctx context.Context, baseURL *url.URL, periodURLs []*url.URL, delay time.Duration) ([]*url.URL, error) {
	sessionURLs := map[string]bool{}
	for _, period := range periodURLs {
		periodPageData, err := RequestPage(ctx, period.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get page %s: %w", period.String(), err)
		}
		periodPage := string(periodPageData)
		sessions := getSessionURLs(periodPage)
		for _, sess := range sessions {
			sessionURLs[sess] = true
		}
		time.Sleep(delay)
	}

	results := []*url.URL{}
	for loc := range sessionURLs {
		pageURL, err := baseURL.Parse(loc)
		if err != nil {
			return nil, fmt.Errorf("failed to construct session page for %s: %w", loc, err)
		}
		results = append(results, pageURL)
	}
	return results, nil
}

func collectVoteFileLocations(ctx context.Context, baseURL url.URL, sessionURLs []*url.URL, delay time.Duration) ([]voteFileURL, error) {
	result := []voteFileURL{}
	for _, sess := range sessionURLs {
		sessionPageData, err := RequestPage(ctx, sess.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get session page %s, %w", sess.String(), err)
		}
		sessionPage := string(sessionPageData)
		iVote, pVote := getVotingURLs(sessionPage)
		individualURL, _ := baseURL.Parse(iVote)
		partyURL, _ := baseURL.Parse(pVote)
		result = append(result, voteFileURL{
			SessionURL:    sess.String(),
			MemberVoteURL: individualURL.String(),
			PartyVoteURL:  partyURL.String(),
		})
		time.Sleep(delay)
	}
	return result, nil
}

func createVoteFiles(ctx context.Context, parliamentURL *url.URL, saveFileName string, log *logrus.Logger) {
	logEntry := log.WithFields(map[string]interface{}{
		"baseURL":  parliamentURL.String(),
		"saveFile": saveFileName,
	})

	logEntry.Info("Start scraping")

	sessionPeriodURLs, err := collectAllSessionPeriods(ctx, parliamentURL)
	if err != nil {
		logEntry.WithError(err).Errorf("failed to collect session period urls")
		return
	}

	sessionURLs, err := collectSessionURLs(ctx, parliamentURL, sessionPeriodURLs, time.Second)
	if err != nil {
		logEntry.WithError(err).Errorf("failed to collect session urls")
		return
	}

	voteFileLoc, err := collectVoteFileLocations(ctx, *parliamentURL, sessionURLs, time.Second*5)
	if err != nil {
		logEntry.WithError(err).Error("failed to collect vote file locations")
		return
	}

	f, err := os.Create(saveFileName)
	if err != nil {
		logEntry.WithError(err).Fatal("failed to create file")
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(voteFileLoc)
	if err != nil {
		logEntry.WithError(err).Fatal("failed to marshal voting data")
	}
}

func getSessionPeriodURLs(body string) []string {
	re := regexp.MustCompile(`/bg/plenaryst/ns/\d+/period/[\d-]+`)
	periods := re.FindAllString(body, -1)
	return periods
}

func getSessionURLs(body string) []string {
	re := regexp.MustCompile(`/bg/plenaryst/ns/\d+/ID/\d+`)
	sessions := re.FindAllString(body, -1)
	return sessions
}

func getVotingURLs(body string) (string, string) {

	pvRE := regexp.MustCompile(`/pub/StenD/\d+gv\d+\.xls`)
	ivRE := regexp.MustCompile(`/pub/StenD/\d+iv\d+\.xls`)
	iVote := ivRE.FindString(body)
	pVote := pvRE.FindString(body)

	return iVote, pVote
}
