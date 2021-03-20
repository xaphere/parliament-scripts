package collectors

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/xaphere/parlament-scripts/pkg/parliament/models"
)

func getSessionUIDs(body string) []string {
	re := regexp.MustCompile(`/bg/plenaryst/ns/\d+/ID/\d+`)
	sessions := re.FindAllString(body, -1)
	result := []string{}
	for _, s := range sessions {
		result = append(result, s[strings.LastIndex(s, "/")+1:])
	}
	return result
}

func GatherPlenarySessionUIDs(ctx context.Context, base string, parliamentID int, periodStart, periodEnd time.Time) ([]string, error) {
	baseLoc, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	result := []string{}
	for now := periodStart; now.Before(periodEnd); now = now.AddDate(0, 1, 0) {
		pageLoc, err := baseLoc.Parse(fmt.Sprintf(`/bg/plenaryst/ns/%d/period/%s`, parliamentID, now.Format("2006-01")))
		if err != nil {
			return nil, fmt.Errorf("failed to construct page address for ID %d, time %s: %w", parliamentID, now.String(), err)
		}
		data, err := RequestPageData(ctx, pageLoc.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get page %s: %w", pageLoc.String(), err)
		}
		result = append(result, getSessionUIDs(string(data))...)
	}
	return result, nil
}

func ExtractProceedingData(proceedingURL *url.URL, reader io.Reader) (*models.Proceeding, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse data: %w", err)
	}
	contentDOM := doc.Find("#leftcontent-2")
	headerDOM := contentDOM.Find(".marktitle")
	timeDOM := headerDOM.Find(".dateclass")
	created, err := time.Parse("02/01/2006", timeDOM.Text())
	if err != nil {
		return nil, err
	}
	attachments := []*url.URL{}
	contentDOM.Find(".markframe .frontList").Find("a").Each(func(i int, sel *goquery.Selection) {
		href, ok := sel.Attr("href")
		if !ok {
			return
		}
		u, err := url.Parse(href)
		if err != nil {
			return
		}
		attachments = append(attachments, proceedingURL.ResolveReference(u))
	})

	title := headerDOM.Text()
	if idx := strings.Index(title, "\n"); idx != -1 {
		title = title[:idx]
	}

	transcriptDOM := contentDOM.Find(".markcontent")
	transcript, err := transcriptDOM.Html()
	if err != nil {
		transcript = transcriptDOM.Text()
	}

	p := &models.Proceeding{
		UID:         getProceedingIDFromURL(proceedingURL),
		Name:        title,
		Date:        created,
		URL:         proceedingURL,
		Transcript:  transcript,
		Attachments: attachments,
		Program:     nil,
		Votes:       nil,
	}
	return p, nil
}

func getProceedingIDFromURL(loc *url.URL) models.ProceedingID {
	str := loc.String()
	idx := strings.LastIndex(str, "/")
	return models.ProceedingID(str[idx+1:])
}

func getVoteFromString(proceedingID string, data string) (*models.Vote, error) {
	re := regexp.MustCompile(`Номер \((?P<id>\d+)\) (?P<type>\p{L}+) проведено на (?P<date>[\d\s:-]+) по тема (?P<title>.*)`)
	const template = `$id|$type|$date|$title`
	result := []byte{}
	submatch := re.FindAllStringSubmatchIndex(data, -1)
	if len(submatch) != 1 {
		return nil, errors.New("no matches found")
	}
	extracted := re.ExpandString(result, template, data, submatch[0])
	str := strings.Split(string(extracted), "|")
	if len(str) != 4 {
		return nil, errors.New("failed to extract valid data")
	}
	date, err := time.Parse(`02-01-2006 15:04`, str[2])
	if err != nil {
		return nil, err
	}
	return &models.Vote{
		UID:   models.VoteID(proceedingID + "-" + str[0]),
		Date:  date,
		Title: str[3],
	}, nil
}
