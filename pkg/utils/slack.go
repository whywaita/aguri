package utils

import (
	"context"
	"fmt"
	"regexp"
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

// ConvertIDToNameInMsg convert channel name in msg
func ConvertIDToNameInMsg(ctx context.Context, msg string, ev *slack.MessageEvent, fromAPI *slack.Client) (string, error) {
	userIds := reUser.FindAllStringSubmatch(ev.Text, -1)
	if len(userIds) != 0 {
		for _, ids := range userIds {
			id := strings.TrimPrefix(ids[0], "<@")
			id = strings.TrimSuffix(id, ">")
			name, _, err := ConvertDisplayUserName(ctx, fromAPI, ev, id)
			if err != nil {
				return "", err
			}
			msg = strings.Replace(msg, id, name, -1)
		}
	}

	return msg, nil
}

// GetUserInfo get info of user
func GetUserInfo(ctx context.Context, fromAPI *slack.Client, ev *slack.MessageEvent) (username, icon string, err error) {
	// get source username and channel, im, group
	user, usertype, err := ConvertDisplayUserName(ctx, fromAPI, ev, "")
	if err != nil {
		return "", "", fmt.Errorf("failed to convert display name: %w", err)
	}

	if usertype == "user" {
		u, err := fromAPI.GetUserInfo(ev.Msg.User)
		if err != nil {
			return "", "", fmt.Errorf("failed to get user info: %w", err)
		}
		icon = u.Profile.Image192
	} else {
		icon = ""
	}

	return user, icon, nil
}

// PostMessageToChannel port message to aggrChannelName
func PostMessageToChannel(ctx context.Context, toAPI, fromAPI *slack.Client, ev *slack.MessageEvent, msg, aggrChannelName string) error {
	// post aggregate message
	var err error

	isExist, _, err := IsExistChannel(ctx, toAPI, aggrChannelName)
	if isExist == false {
		return fmt.Errorf("channel is not found: %w", err)
	}
	if err != nil {
		return fmt.Errorf("failed to get info of exist channel: %w", err)
	}

	user, icon, err := GetUserInfo(ctx, fromAPI, ev)
	fType, position, err := ConvertDisplayChannelName(ctx, fromAPI, ev)
	if err != nil {
		return fmt.Errorf("failed to convert channel name: %w", err)
	}

	param := slack.PostMessageParameters{
		IconURL: icon,
	}
	username := user + "@" + strings.ToLower(fType[:1]) + ":" + position
	if ev.ThreadTimestamp != "" {
		username += " (in Thread)"
	}
	param.Username = username

	attachments := ev.Attachments

	// convert user id to username in message
	msg, err = ConvertIDToNameInMsg(ctx, msg, ev, fromAPI)
	if err != nil {
		return fmt.Errorf("failed to convert id to name: %w", err)
	}

	workspace := strings.TrimPrefix(aggrChannelName, config.PrefixSlackChannel)
	if msg != "" {
		respChannel, respTimestamp, err := toAPI.PostMessageContext(ctx, aggrChannelName, slack.MsgOptionText(msg, true), slack.MsgOptionPostMessageParameters(param))
		if err != nil {
			return fmt.Errorf("failed to post message: %w", err)
		}
		store.SetSlackLog(workspace, ev.Timestamp, position, msg, respChannel, respTimestamp)
	}
	// if msg is blank, maybe bot_message (for example, twitter integration).
	// so, must post blank msg if this post has attachments.
	if attachments != nil {
		for _, attachment := range attachments {
			respChannel, respTimestamp, err := toAPI.PostMessageContext(ctx, aggrChannelName, slack.MsgOptionPostMessageParameters(param), slack.MsgOptionAttachments(attachment))
			if err != nil {
				return fmt.Errorf("failed to post message: %w", err)
			}
			store.SetSlackLog(workspace, ev.Timestamp, position, msg, respChannel, respTimestamp)
		}
	}

	return nil
}

// GenerateAguriUsername generate name that format of aguri
func GenerateAguriUsername(ch *slack.Channel, displayUsername string) string {
	return displayUsername + "@" + strings.ToLower(ch.ID[:1]) + ":" + ch.Name
}
