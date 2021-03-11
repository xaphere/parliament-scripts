package main

import (
	"context"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"

	"github.com/xaphere/parlament-scripts/pkg/parliament/collectors"
	"github.com/xaphere/parlament-scripts/pkg/parliament/models"
	"github.com/xaphere/parlament-scripts/pkg/parliament/storage"
	"github.com/xaphere/parlament-scripts/pkg/scripts"
)

func main() {
	log := logrus.New()

	ids, err := scripts.ExtractMPIDs()
	if err != nil {
		log.WithError(err).Fatal("failed foo")
	}
	ctx := context.Background()
	for _, id := range ids {
		scripts.GetMPPage(ctx, id)
	}
}

func work(log *logrus.Logger) {
	pageURL, err := url.Parse("https://www.parliament.bg/bg/plenaryst/ns/52/ID/10474")
	if err != nil {
		log.WithError(err).Fatal("failed to parse page url")
	}

	resp, err := http.Get(pageURL.String())
	if err != nil {
		log.WithError(err).Fatal("failed to get proceedings page")
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("unexpected status code %s", http.StatusText(resp.StatusCode))
	}
	defer resp.Body.Close()
	p, err := collectors.ExtractProceedingData(pageURL, resp.Body)
	if err != nil {
		log.WithError(err).Fatal("failed to create proceedings object")
	}

	db, err := storage.NewDBConnection("postgres://postgres:postgres@localhost:5432")
	if err != nil {
		log.WithError(err).Fatal("failed to connect to database")
	}
	ctx := context.Background()
	initPrailiamnetData(ctx, db, log)
	db.CreateProceeding(ctx, p)

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
