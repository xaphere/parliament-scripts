package scripts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type voteFileURL struct {
	SessionURL    string `json:"session"`
	MemberVoteURL string `json:"named_vote"`
	PartyVoteURL  string `json:"party_vote"`
}

func ReadVotesFile(fileLoc string) ([]voteFileURL, error) {
	data := []voteFileURL{}
	voteFile, err := os.Open(fileLoc)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer voteFile.Close()
	err = json.NewDecoder(voteFile).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file : %w", err)
	}
	return data, nil
}

func downloadVoteFiles(ctx context.Context, voteFileLoc string, storageLoc string, log *logrus.Logger) {
	logEntry := log.WithField("voteFile", voteFileLoc)
	voteLoc, err := ReadVotesFile(voteFileLoc)
	if err != nil {
		logEntry.WithError(err).Error("failed to get votes data")
		return
	}

	shouldDownloadFile := func(fileURL string) bool {
		fileName := GetFileNameFromURL(fileURL)
		return fileURL != "" && !fileExists(filepath.Join(storageLoc, fileName))
	}

	for _, loc := range voteLoc {

		if !shouldDownloadFile(loc.MemberVoteURL) && !shouldDownloadFile(loc.PartyVoteURL) {
			continue
		}
		logEntry := logEntry.WithFields(map[string]interface{}{
			"member":  loc.MemberVoteURL,
			"party":   loc.PartyVoteURL,
			"session": loc.SessionURL,
		})
		logEntry.Info("downloading files")

		if shouldDownloadFile(loc.MemberVoteURL) {
			err = downloadFile(ctx, loc.MemberVoteURL, storageLoc)
			if err != nil {
				logEntry.WithError(err).Error("failed to download member file")
			}
		}

		if shouldDownloadFile(loc.PartyVoteURL) {
			err = downloadFile(ctx, loc.PartyVoteURL, storageLoc)
			if err != nil {
				logEntry.WithError(err).Error("failed to download party file")
			}
		}
	}

}

func downloadFile(ctx context.Context, fileURL string, storage string) error {

	fileName := GetFileNameFromURL(fileURL)
	if filepath.Ext(fileName) == "" {
		return fmt.Errorf("can't create file name from %s", fileURL)
	}

	data, err := RequestPage(ctx, fileURL)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(storage, fileName))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func fileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func GetFileNameFromURL(fileURL string) string {
	slashIdx := strings.LastIndex(fileURL, "/")
	fileName := fileURL[slashIdx+1:]
	return fileName
}

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

func RequestPage(ctx context.Context, page string) ([]byte, error) {
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
		return RequestPage(ctx, page)
	}
	return data, nil
}
