package main

import (
	"context"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/xaphere/parlament-scripts/pkg/parliament/collectors"
	"github.com/xaphere/parlament-scripts/pkg/parliament/storage"
)

func main() {
	log := logrus.New()

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

	db := storage.NewDB("postgres://postgres:postgres@localhost:5432")
	//db, err := storage.NewStorage("host=localhost user=postgres password=postgres DB.name=postgres port=5432")
	ctx := context.Background()
	err = db.Connect(ctx)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to database")
	}
	defer db.Disconnect(ctx)
	err = db.CreateProceeding(ctx, p)
	if err != nil {
		log.WithError(err).Fatal("failed to store proceeding")
	}

	_, err = db.ReadProceeding(ctx, p.UID)
	if err != nil {
		log.WithError(err).Fatal("failed to read proceeding")
	}
}
