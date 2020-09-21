package main

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
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

func requestPage(ctx context.Context, page string) ([]byte, error) {
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
		return requestPage(ctx, page)
	}
	return data, nil
}
