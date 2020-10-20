package scripts

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	data, err := requestPage(ctx, fileURL)
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
