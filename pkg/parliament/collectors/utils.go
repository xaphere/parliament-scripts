package collectors

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/xaphere/parlament-scripts/pkg/parliament/models"
)

type deltaKeyType string

const deltaKey deltaKeyType = "delta"
const deltaIncrement = 30
const deltaInit = 5

func newContextWithDelta(ctx context.Context, delta int) context.Context {
	return context.WithValue(ctx, deltaKey, delta)
}
func deltaFromContext(ctx context.Context) (int, bool) {
	delta, ok := ctx.Value(deltaKey).(int)
	return delta, ok
}

func RequestPageData(ctx context.Context, page string) ([]byte, error) {
	if _, ok := deltaFromContext(ctx); !ok {
		ctx = newContextWithDelta(ctx, deltaInit)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, page, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	pageData := string(data)
	if strings.Contains(pageData, "User validation required to continue..") {
		//we've hit the ddos protection of the site. let's wait for a bit and try again
		delta := ctx.Value(deltaKey).(int)
		if delta > deltaIncrement*10 {
			return nil, errors.New("stuck in validation")
		}
		time.Sleep(time.Second * time.Duration(delta))
		ctx = context.WithValue(ctx, deltaKey, delta+deltaIncrement)
		return RequestPageData(ctx, page)
	}
	return data, nil
}

func GetPageReader(ctx context.Context, page string) (io.Reader, error) {
	data, err := RequestPageData(ctx, page)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(data)
	return reader, nil
}

// GetLocalPartyID uses constants to resolve partyID from string
func GetLocalPartyID(party string) (int, error) {
	parties := map[int]string{
		2544: "герб",
		2545: "бсп",
		2547: "дпс",
		2546: "обединени патриоти",
		2549: "нечленуващи в пг",
		2548: "воля",
	}
	cmpParty := strings.ToLower(party)
	for id, name := range parties {
		if strings.Contains(cmpParty, name) {
			return id, nil
		}
	}
	return 0, errors.New("not found")
}

// GetLocalConstituencyID uses constants to resolve constituencyID from string
func GetLocalConstituencyID(constituency string) (int, error) {
	constituencies := []models.Constituency{
		{1, "Благоевград"},
		{2, "Бургас"},
		{3, "Варна"},
		{4, "Велико Търново"},
		{5, "Видин"},
		{6, "Враца"},
		{7, "Габрово"},
		{8, "Добрич"},
		{9, "Кърджали"},
		{10, "Кюстендил"},
		{11, "Ловеч"},
		{12, "Монтана"},
		{13, "Пазарджик"},
		{14, "Перник"},
		{15, "Плевен"},
		{16, "Пловдив град"},
		{17, "Пловдив област"},
		{18, "Разград"},
		{19, "Русе"},
		{20, "Силистра"},
		{21, "Сливен"},
		{22, "Смолян"},
		{23, "СОФИЯ 1"},
		{24, "СОФИЯ 2"},
		{25, "СОФИЯ 3"},
		{26, "София област"},
		{27, "Стара Загора"},
		{28, "Търговище"},
		{29, "Хасково"},
		{30, "Шумен"},
		{31, "Ямбол"},
		{32, "Чужбина"},
	}
	cmpCon := strings.ToLower(constituency)
	for _, c := range constituencies {
		cmpName := strings.ToLower(c.Name)
		if strings.Contains(cmpCon, cmpName) {
			return c.ID, nil
		}
	}
	return 0, errors.New("not found")
}
