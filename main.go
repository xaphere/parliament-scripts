package main

import (
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/xaphere/parliament/collector"
)

const pageURL = "https://www.parliament.bg/bg/plenaryst/ns/52/ID/10474"

func main() {
	log := logrus.New()

	//resp, err := http.Get("https://www.parliament.bg/bg/plenaryst/ns/52/ID/10474")
	//if err != nil {
	//	log.WithError(err).Fatal("failed to get proceedings page")
	//}
	//defer resp.Body.Close()
	//data, err := ioutil.ReadAll(resp.Body)
	u, err := url.Parse(pageURL)
	if err != nil {
		log.WithError(err).Fatal("failed to read response body")
	}
	_, err = collector.ExtractData(u)
	if err != nil {
		log.WithError(err).Fatal("failed to create proceedings object")
	}
}
