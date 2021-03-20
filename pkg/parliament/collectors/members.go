package collectors

import (
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/xaphere/parlament-scripts/pkg/parliament/models"
)

const parliamentAddress = "https://www.parliament.bg/"
const parliamentMPAddress = parliamentAddress + "bg/MP/"

// resolver deduces objectID by the object name
type resolver func(name string) (int, error)

func ExtractMemberIDs(reader io.Reader) ([]string, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	content := string(data)
	pattern := regexp.MustCompile(`\/bg\/MP\/(?P<id>\d+)`)
	template := `$id `
	result := []byte{}
	for _, submatches := range pattern.FindAllStringSubmatchIndex(content, -1) {
		result = pattern.ExpandString(result, template, content, submatches)
	}
	memberIds := strings.Fields(string(result))
	return memberIds, nil
}

func ExtractMember(reader io.Reader, memberID int, toParty, toConstituency resolver) (*models.Member, error) {
	var (
		name           = ""
		email          = ""
		partyID        = -1
		constituencyID = -1
		err            error
	)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}
	name = doc.Find(".MProwD").Text()
	doc.Find(".MPinfo li").Each(func(i int, sel *goquery.Selection) {
		body := sel.Text()
		switch {
		case strings.HasPrefix(body, "Избран(а) с политическа сила: "):
			text := strings.TrimPrefix(body, "Избран(а) с политическа сила: ")
			partyID, err = toParty(text)
			if err != nil {
				return
			}
		case strings.HasPrefix(body, "Изборен район: "):
			text := strings.TrimPrefix(body, "Изборен район: ")
			constituencyID, err = toConstituency(text)
			if err != nil {
				return
			}
		case strings.HasPrefix(body, "E-mail: "):
			email = strings.TrimPrefix(body, "E-mail: ")
		}

	})
	if partyID < 2 || constituencyID < 1 {
		return nil, fmt.Errorf("no full data for %d, party %d constituency %d", memberID, partyID, constituencyID)
	}
	return &models.Member{
		ID:             memberID,
		Name:           name,
		PartyID:        partyID,
		ConstituencyID: constituencyID,
		Email:          email,
	}, nil
}
