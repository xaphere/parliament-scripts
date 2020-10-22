package collector

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/xaphere/parliament/models"
)

func ExtractData(proceedingURL url.URL) (*models.Proceeding, error) {

	resp, err := http.Get(proceedingURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	p, err := extractData(resp.Body)
	if err != nil {
		return nil, err
	}
	p.URL = proceedingURL
	p.UID = getProceedingIDFromURL(proceedingURL)
	return p, nil
}

func getProceedingIDFromURL(loc url.URL) models.ProceedingID {
	str := loc.String()
	idx := strings.LastIndex(str, "/")
	return models.ProceedingID(str[idx+1:])
}

func extractData(reader io.Reader) (*models.Proceeding, error) {

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
	attachments := []url.URL{}
	contentDOM.Find(".markframe .frontList").Find("a").Each(func(i int, sel *goquery.Selection) {
		href, ok := sel.Attr("href")
		if !ok {
			return
		}
		u, err := url.Parse(href)
		if err != nil {
			return
		}
		attachments = append(attachments, *u)
	})

	transcriptDOM := contentDOM.Find(".markcontent")

	p := &models.Proceeding{
		UID:         "",
		Name:        titleDOM.Text(),
		Date:        created,
		URL:         url.URL{},
		Transcript:  transcriptDOM.Text(),
		Attachments: attachments,
		ProgramID:   "",
		Votes:       nil,
	}
	return p, nil
}
