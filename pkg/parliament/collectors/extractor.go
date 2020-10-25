package collectors

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/xaphere/parlament-scripts/pkg/parliament/models"

	"github.com/PuerkitoBio/goquery"
)

func ExtractProceedingData(reader io.Reader) (*models.Proceeding, error) {

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse data: %w", err)
	}
	contentDOM := doc.Find("#leftcontent-2")
	titleDOM := contentDOM.Find(".marktitle")
	timeDOM := titleDOM.Find(".dateclass")
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
		attachments = append(attachments, u)
	})

	transcriptDOM := contentDOM.Find(".markcontent")

	p := &models.Proceeding{
		UID:         getProceedingIDFromURL(proceedingURL),
		Name:        titleDOM.Text(),
		Date:        created,
		URL:         proceedingURL,
		Transcript:  transcriptDOM.Text(),
		Attachments: attachments,
		ProgID:      "",
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