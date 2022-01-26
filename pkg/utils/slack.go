package utils

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackutilsx"
	"github.com/whywaita/aguri/pkg/config"
	"github.com/whywaita/aguri/pkg/store"
)

var (
	reUser = regexp.MustCompile(`<@U(\S+)>`)
)

// IsExistChannel check exist
func IsExistChannel(ctx context.Context, api *slack.Client, searchName string) (bool, *slack.Channel, error) {
	// channel is exist => True
	channels, err := GetConversationsList(ctx, api, []slackutilsx.ChannelType{slackutilsx.CTypeChannel, slackutilsx.CTypeGroup})
	if err != nil {
		return false, nil, fmt.Errorf("failed to get conversation list: %w", err)
	}

	for _, channel := range channels {
		if channel.Name == searchName {
			return true, &channel, nil
		}
	}
	return false, nil, fmt.Errorf("%s is not found", searchName)
}

// GetMessageByTS get message history from API
func GetMessageByTS(ctx context.Context, api *slack.Client, channel, timestamp string) (*slack.Message, error) {
	// get message via RestAPI by Timestamp
	// want to get only one message
	historyParam := &slack.GetConversationHistoryParameters{
		ChannelID: channel,
		Latest:    timestamp,
		Oldest:    timestamp,
	}

	history, err := api.GetConversationHistoryContext(ctx, historyParam)
	if err != nil {
		return nil, fmt.Errorf("failed to get message history by timestamp: %w", err)
	}

	msg := history.Messages[0]

	return &msg, nil
}

func getInitial(conversationType slackutilsx.ChannelType) string {
	str := conversationType.String()
	return strings.ToLower(str[:1])
}

func getAggredUsername(username string, conversationType slackutilsx.ChannelType, conversationName string, isThead bool) string {
	aggredUsername := username + "@" + getInitial(conversationType) + ":" + conversationName
	if isThead {
		aggredUsername += " (in Thread)"
	}
	return aggredUsername
}

func getPostParamMessageEvent(ctx context.Context, fromAPI *slack.Client, ev *slack.MessageEvent) (string, string, slackutilsx.ChannelType, string, error) {
	username, _, iconURL, err := GetUserNameTypeIconMessageEvent(ctx, fromAPI, ev)
	if err != nil {
		return "", "", slackutilsx.CTypeUnknown, "", fmt.Errorf("failed to get user info: %w", err)
	}
	fromType, conversationName, err := ConvertDisplayChannelNameMessageEvent(ctx, fromAPI, ev)
	if err != nil {
		return "", "", slackutilsx.CTypeUnknown, "", fmt.Errorf("failed to convert channel name: %w", err)
	}

	return username, iconURL, fromType, conversationName, nil
}

// PostMessageToChannelMessageEvent port message to aggrConversationName from slack.MessageEvent
func PostMessageToChannelMessageEvent(ctx context.Context, toAPI, fromAPI *slack.Client, ev *slack.MessageEvent, msg, aggrConversationName string) error {
	username, iconURL, fromType, conversationName, err := getPostParamMessageEvent(ctx, fromAPI, ev)
	if err != nil {
		return fmt.Errorf("failed to get param: %w", err)
	}

	isThread := ev.Msg.ThreadTimestamp != ""
	aggredUsername := getAggredUsername(username, fromType, conversationName, isThread)

	return PostMessageToChannel(ctx,
		toAPI,
		iconURL,
		aggredUsername,
		ev.Attachments,
		conversationName,
		ev.Timestamp,
		msg,
		aggrConversationName,
	)
}

// PostMessageToChannelUploadedFile port file link to aggrConversationName from slack.FileSharedEvent
func PostMessageToChannelUploadedFile(ctx context.Context, toAPI, fromAPI *slack.Client, ev *slack.FileSharedEvent, originalFile, uploadedFile *slack.File, aggrConversationName string) error {
	username, iconURL, fromType, conversationName, err := getPostParam(ctx, fromAPI, originalFile.User, ev.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to get param: %w", err)
	}

	sharedFileInfo := isSharedFile(originalFile, ev.ChannelID)
	if sharedFileInfo == nil {
		return fmt.Errorf("failed to get shared info from file: %w", err)
	}
	newestSharedFileInfo := sharedFileInfo[0]
	isThread := newestSharedFileInfo.ThreadTs != ""
	aggredUsername := getAggredUsername(username, fromType, conversationName, isThread)

	attachments := []slack.Attachment{
		{
			// todo: Permalink will not unfurl, fix me
			Text: uploadedFile.Permalink,
		},
	}

	return PostMessageToChannel(ctx,
		toAPI,
		iconURL,
		aggredUsername,
		attachments,
		conversationName,
		ev.EventTimestamp,
		"",
		aggrConversationName,
	)
}

func isSharedFile(f *slack.File, sharedChannelID string) []slack.ShareFileInfo {
	infoPublic, ok := f.Shares.Public[sharedChannelID]
	if ok {
		sort.SliceStable(infoPublic, func(i, j int) bool {
			return infoPublic[i].Ts > infoPublic[j].Ts
		})
		return infoPublic
	}

	infoPrivate, ok := f.Shares.Private[sharedChannelID]
	if ok {
		sort.SliceStable(infoPrivate, func(i, j int) bool {
			return infoPrivate[i].Ts > infoPrivate[j].Ts
		})
		return infoPrivate
	}

	return nil
}

func getPostParam(ctx context.Context, fromAPI *slack.Client, originalUserID, originalChannelID string) (string, string, slackutilsx.ChannelType, string, error) {
	username, _, iconURL, err := getUserNameTypeIconFileSharedEvent(ctx, fromAPI, originalUserID)
	if err != nil {
		return "", "", slackutilsx.CTypeUnknown, "", fmt.Errorf("failed to get user info: %w", err)
	}

	fromType, conversationName, err := ConvertDisplayChannelName(ctx, fromAPI, originalChannelID, originalUserID, "")
	if err != nil {
		return "", "", slackutilsx.CTypeUnknown, "", fmt.Errorf("failed to convert channel name: %w", err)
	}

	return username, iconURL, fromType, conversationName, nil
}

// PostMessageToChannel port message to aggrConversationName
func PostMessageToChannel(
	ctx context.Context,
	toAPI *slack.Client,
	iconURL string,
	aggredUsername string,
	attachments []slack.Attachment,
	fromConversationName string,
	fromTimestamp string,
	msg, aggrConversationName string,
) error {
	// post aggregate message
	var err error

	isExist, _, err := IsExistChannel(ctx, toAPI, aggrConversationName)
	if isExist == false {
		return fmt.Errorf("channel is not found: %w", err)
	}
	if err != nil {
		return fmt.Errorf("failed to get info of exist channel: %w", err)
	}

	param := slack.PostMessageParameters{
		IconURL:     iconURL,
		Username:    aggredUsername,
		UnfurlMedia: true,
	}

	workspace := strings.TrimPrefix(aggrConversationName, config.PrefixSlackChannel)
	if msg != "" {
		respChannel, respTimestamp, err := toAPI.PostMessageContext(
			ctx,
			aggrConversationName,
			slack.MsgOptionText(msg, true),
			slack.MsgOptionPostMessageParameters(param),
		)
		if err != nil {
			return fmt.Errorf("failed to post message: %w", err)
		}
		store.SetSlackLog(workspace, fromTimestamp, fromConversationName, msg, respChannel, respTimestamp)
	}
	// if msg is blank, maybe bot_message (for example, twitter integration).
	// so, must post blank msg if this post has attachments.
	if attachments != nil {
		for _, attachment := range attachments {
			respChannel, respTimestamp, err := toAPI.PostMessageContext(
				ctx,
				aggrConversationName,
				slack.MsgOptionPostMessageParameters(param),
				slack.MsgOptionAttachments(attachment),
			)
			if err != nil {
				return fmt.Errorf("failed to post message: %w", err)
			}
			store.SetSlackLog(workspace, fromTimestamp, fromConversationName, msg, respChannel, respTimestamp)
		}
	}

	return nil
}

// GenerateAguriUsername generate name that format of aguri
func GenerateAguriUsername(ch *slack.Channel, displayUsername string) string {
	return displayUsername + "@" + strings.ToLower(ch.ID[:1]) + ":" + ch.Name
}
