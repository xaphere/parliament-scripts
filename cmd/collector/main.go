package main

import (
	"bytes"
	"context"
	"net/url"
	"time"

	_ "github.com/lib/pq"

	"github.com/sirupsen/logrus"

	"github.com/xaphere/parlament-scripts/pkg/parliament/collectors"
	"github.com/xaphere/parlament-scripts/pkg/parliament/models"
	"github.com/xaphere/parlament-scripts/pkg/parliament/storage"
	"github.com/xaphere/parlament-scripts/pkg/scripts"
)

const parliamentLocation = "https://www.parliament.bg/"
const currentParliamentID = 52
const parliamentStartDate = "2017-04-01T15:04:05Z"

func main() {
	log := logrus.New()
	db, err := storage.NewDBConnection("postgres://postgres:postgres@localhost:5432")
	if err != nil {
		log.WithError(err).Fatal("failed to connect to database")
	}
	ctx := context.Background()
	//initPrailiamnetData(ctx, db, log)
	sessionIds, err := GetPlenarySessionUIDs(ctx)
	if err != nil {
		log.WithError(err).Fatal()
	}
	err = gatherSessions(ctx, sessionIds, db)
	if err != nil {
		log.WithError(err).Fatal()
	}
}

func GetPlenarySessionUIDs(ctx context.Context) ([]string, error) {
	start, err := time.Parse(time.RFC3339, parliamentStartDate)
	if err != nil {
		return nil, err
	}
	end := start.AddDate(4, 0, 0)
	return collectors.GatherPlenarySessionUIDs(context.Background(), parliamentLocation, currentParliamentID, start, end)
}

func downloader() {
	ctx := context.Background()
	log := logrus.New()

	db, err := storage.NewDBConnection("postgres://postgres:postgres@localhost:5432")
	if err != nil {
		log.WithError(err).Fatal("failed to connect to database")
	}

	members, err := collectors.GetMembers(ctx)
	if err != nil {
		log.WithError(err).Fatal("failed to get members")
	}

	for _, member := range members {

		logEntry := log.WithField("memberID", member.ID)
		err = db.CreateMember(ctx, *member)
		if err != nil {
			logEntry.WithError(err).Error("failed to store member")
		}
	}
}

func gatherSessions(ctx context.Context, sessionIds []string, db *storage.SQLStorage) error {
	for _, id := range sessionIds {
		pageURL, err := url.Parse("https://www.parliament.bg/bg/plenaryst/ns/52/ID/" + id)
		if err != nil {
			return err
		}

		resp, err := scripts.RequestPage(ctx, pageURL.String())
		if err != nil {
			return err
		}

		p, err := collectors.ExtractProceedingData(pageURL, bytes.NewReader(resp))
		if err != nil {
			return err
		}
		err = db.CreateProceeding(ctx, p)
		if err != nil {
			return err
		}
	}
	return nil
}

func initPrailiamnetData(ctx context.Context, db *storage.SQLStorage, log *logrus.Logger) {
	parties := []models.Party{
		{2544, "ГЕРБ"},
		{2545, "БСП за България"},
		{2547, "Движение за права и свободи"},
		{2546, "Обединени патриоти"},
		{2549, "Нечленуващи в ПГ"},
		{2548, "ВОЛЯ - Българските Родолюбци"},
	}

	for _, party := range parties {
		err := db.CreateParty(ctx, party)
		if err != nil {
			log.WithError(err).WithField("party", party.Name).Fatal("failed to create party")
		}
	}
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
		{23, "София 23 МИР1"},
		{24, "София 24 МИР1"},
		{25, "София 25 МИР1"},
		{26, "София област"},
		{27, "Стара Загора"},
		{28, "Търговище"},
		{29, "Хасково"},
		{30, "Шумен"},
		{31, "Ямбол"},
		{32, "Чужбина"},
	}
	for _, c := range constituencies {
		err := db.CreateConstituency(ctx, c)
		if err != nil {
			log.WithError(err).WithField("constituency", c.Name).Fatal("failed to create Constituency")
		}
	}
}
