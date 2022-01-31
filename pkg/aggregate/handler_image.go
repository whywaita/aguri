package aggregate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/slack-go/slack/slackutilsx"

	"github.com/whywaita/aguri/pkg/config"
	"github.com/whywaita/aguri/pkg/utils"

	"github.com/whywaita/aguri/pkg/store"

	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

func handleFileSharedEvent(ctx context.Context, ev *slack.FileSharedEvent, fromAPI *slack.Client, workspace, lastTimestamp string, logger *logrus.Logger) string {
	f, _, _, err := fromAPI.GetFileInfoContext(ctx, ev.File.ID, 100, 0)
	if err != nil {
		logger.Warn(err)
		return ev.EventTimestamp
	}

	joinedChannels, err := utils.GetJoinedConversationsList(ctx, fromAPI, []slackutilsx.ChannelType{
		slackutilsx.CTypeChannel, slackutilsx.CTypeGroup, slackutilsx.CTypeDM,
	})
	if err != nil {
		logger.Warn(err)
		return ev.EventTimestamp
	}
	if joined := utils.IsJoined(ev.ChannelID, joinedChannels); !joined {
		// not upload if not joined channel
		return ev.EventTimestamp
	}

	fileByte, err := downloadFile(ctx, f.URLPrivateDownload, workspace)
	if err != nil {
		logger.Warn(err)
		return ev.EventTimestamp
	}

	uploadedFile, err := uploadFileWithRetry(ctx, fileByte, f, logger)
	if err != nil {
		logger.Warnf("failed to upload file with retry: %+v", err)
		return ev.EventTimestamp
	}

	if err := utils.PostMessageToChannelUploadedFile(ctx, store.GetConfigToAPI(), fromAPI, ev, f, uploadedFile, config.GetToChannelName(workspace)); err != nil {
		logger.Warnf("failed to post uploaded file: %+v", err)
		return ev.EventTimestamp
	}

	return ev.EventTimestamp
}

func uploadFileWithRetry(ctx context.Context, input []byte, originalFile *slack.File, logger *logrus.Logger) (*slack.File, error) {
	param := slack.FileUploadParameters{
		Filetype:       originalFile.Filetype,
		Filename:       originalFile.Name,
		Title:          originalFile.Title,
		InitialComment: originalFile.InitialComment.Comment,
		// Channels will share channels, but it can't configure some parameter (e.g. username),
		// So `files.upload` is not configure Channel and after share post it.
		//Channels:       []string{config.GetToChannelName(workspace)},
	}

	for i := 0; i < 3; i++ {
		param.Reader = bytes.NewBuffer(input)
		uploadedFile, err := store.GetConfigToAPI().UploadFileContext(ctx, param)
		if err == nil {
			return uploadedFile, nil
		}

		if strings.Contains(err.Error(), "408 Request Timeout") {
			// retryable
			logger.Infof("found upload file is 408 timeout, retry...")
			time.Sleep(1 * time.Second)
			continue
		}
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return nil, fmt.Errorf("failed to upload file 3 times")
}

func downloadFile(ctx context.Context, privateDownloadURL string, workspace string) ([]byte, error) {
	fromAPIToken := store.GetConfigFromAPI(workspace)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, privateDownloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", fromAPIToken))

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return b, nil
}
