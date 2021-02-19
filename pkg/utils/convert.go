package utils

import (
	"github.com/pkg/errors"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackutilsx"
)

const (
	ErrMethodNotSupportedForChannelType = "method_not_supported_for_channel_type"
)

func GetConversationsList(api *slack.Client, types []slackutilsx.ChannelType) ([]slack.Channel, error) {
	var paramTypes []string

	for _, t := range types {
		switch t {
		case slackutilsx.CTypeChannel:
			paramTypes = append(paramTypes, "public_channel")
		case slackutilsx.CTypeGroup:
			paramTypes = append(paramTypes, "private_channel")
		case slackutilsx.CTypeDM:
			paramTypes = append(paramTypes, "im")
		default:
			paramTypes = append(paramTypes, "")
		}
	}

	param := &slack.GetConversationsParameters{
		ExcludeArchived: "true",
		Types:           paramTypes,
	}

	channels, nextCursor, err := api.GetConversations(param)
	if err != nil {
		return nil, err
	}

	for nextCursor != "" {
		param = &slack.GetConversationsParameters{
			ExcludeArchived: "true",
			Types:           paramTypes,
			Cursor:          nextCursor,
		}
		c, n, err := api.GetConversations(param)
		if err != nil {
			return nil, err
		}
		channels = Concat(channels, c)
		nextCursor = n
	}

	return channels, nil
}

func Concat(xs, ys []slack.Channel) []slack.Channel {
	zs := make([]slack.Channel, len(xs)+len(ys))

	for i, x := range xs {
		zs[i] = x
	}

	l := len(xs)
	for j, y := range ys {
		zs[l+j] = y
	}

	return zs
}

func ConvertDisplayPrivateChannel(api *slack.Client, channelId string) (string, error) {
	channels, err := GetConversationsList(api, []slackutilsx.ChannelType{slackutilsx.CTypeGroup})
	if err != nil {
		return "", err
	}

	for _, c := range channels {
		if c.ID == channelId {
			return c.Name, nil
		}
	}

	return "", errors.New("channel not found")
}

func ConvertDisplayChannelName(api *slack.Client, ev *slack.MessageEvent) (fromType, name string, err error) {
	// identify channel or group (as known as private channel) or DM

	channelType := slackutilsx.DetectChannelType(ev.Channel)
	fromType = channelType.String()
	switch channelType {
	case slackutilsx.CTypeChannel:
		info, err := api.GetChannelInfo(ev.Channel)
		if err != nil {
			if err.Error() == ErrMethodNotSupportedForChannelType {
				// This error occurred by the private channels only converted from the public channel.
				// So, this is private channel if this error.
				name, err := ConvertDisplayPrivateChannel(api, ev.Channel)
				if err != nil {
					return "", "", err
				}

				return fromType, name, nil
			} else {
				return "", "", err
			}
		}

		return fromType, info.Name, nil

	case slackutilsx.CTypeGroup:
		info, err := api.GetGroupInfo(ev.Channel)
		if err != nil {
			return "", "", err
		}

		return fromType, info.Name, nil

	case slackutilsx.CTypeDM:
		if ev.Msg.SubType != "" {
			// SubType is not define user
		} else {
			info, err := api.GetUserInfo(ev.Msg.User)
			if err != nil {
				return "", "", err
			}

			return fromType, info.Name, nil
		}
	default:
		name = ""
	}

	return "", "", errors.New("channel not found")
}

func ConvertDisplayUserName(api *slack.Client, ev *slack.MessageEvent, id string) (username, usertype string, err error) {
	// user id to display name

	if id != "" {
		// specific id (maybe user)
		info, err := api.GetUserInfo(id)
		if err != nil {
			return "", "", err
		}

		return info.Name, "user", nil
	}

	// return self id
	if ev.Msg.BotID == "B01" {
		// this is slackbot
		return "Slack bot", "bot", nil
	} else if ev.Msg.BotID != "" {
		// this is bot
		byInfo, err := api.GetBotInfo(ev.Msg.BotID)
		if err != nil {
			return "", "", err
		}

		return byInfo.Name, "bot", nil
	} else if ev.Msg.SubType != "" {
		// SubType is not define user
		return ev.Msg.SubType, "status", nil
	} else {
		// user
		byInfo, err := api.GetUserInfo(ev.Msg.User)
		if err != nil {
			return "", "", err
		}

		return byInfo.Name, "user", nil
	}
}
