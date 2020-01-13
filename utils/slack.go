package utils

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/whywaita/aguri/config"

	"github.com/whywaita/aguri/store"

	"github.com/nlopes/slack"
	"github.com/whywaita/slack_lib"
)

var (
	reUser = regexp.MustCompile(`<@U(\S+)>`)
)

func CheckExistChannel(api *slack.Client, searchName string) (bool, error) {
	// channel is exist => True
	channels, err := api.GetChannels(false)
	if err != nil {
		return false, errors.Wrap(err, "failed to get channel info")
	}

	for _, channel := range channels {
		if channel.Name == searchName {
			// if channel is exist, return true
			return true, nil
		}
	}

	return false, nil
}

func createNewChannel(api *slack.Client, name string) error {
	var err error
	_, err = api.CreateChannel(name)
	if err != nil {
		return errors.Wrap(err, "failed to create new channel")
	}

	return nil
}

func GetMessageByTS(api *slack.Client, channel, timestamp string) (*slack.Message, error) {
	// get message via RestAPI by Timestamp
	// want to get only one message
	historyParam := slack.NewHistoryParameters()
	// historyParam.Count = 1
	historyParam.Latest = timestamp
	historyParam.Oldest = timestamp

	history, err := api.GetChannelHistory(channel, historyParam)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get message history by timestamp")
	}

	msg := history.Messages[0]

	return &msg, nil
}

func ConvertUserIdtoName(msg string, ev *slack.MessageEvent, fromAPI *slack.Client) (string, error) {
	userIds := reUser.FindAllStringSubmatch(ev.Text, -1)
	if len(userIds) != 0 {
		for _, ids := range userIds {
			id := strings.TrimPrefix(ids[0], "<@")
			id = strings.TrimSuffix(id, ">")
			name, _, err := slack_lib.ConvertDisplayUserName(fromAPI, ev, id)
			if err != nil {
				return "", err
			}
			msg = strings.Replace(msg, id, name, -1)
		}
	}

	return msg, nil
}

func GetUserInfo(fromAPI *slack.Client, ev *slack.MessageEvent) (username, icon string, err error) {
	// get source username and channel, im, group
	user, usertype, err := slack_lib.ConvertDisplayUserName(fromAPI, ev, "")
	if err != nil {
		return "", "", errors.Wrap(err, "failed to convert display name")
	}

	if usertype == "user" {
		u, err := fromAPI.GetUserInfo(ev.Msg.User)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to get user info")
		}
		icon = u.Profile.Image192
	} else {
		icon = ""
	}

	return user, icon, nil
}

func PostMessageToChannel(toAPI, fromAPI *slack.Client, ev *slack.MessageEvent, msg, aggrChannelName string) error {
	// post aggregate message
	var err error

	isExist, err := CheckExistChannel(toAPI, aggrChannelName)
	if err != nil {
		return errors.Wrap(err, "failed to get info of exist channel")
	}

	if (isExist == false) && (err == nil) {
		// if channel is not exist, make channel
		err = createNewChannel(toAPI, aggrChannelName)
		if err != nil {
			return errors.Wrap(err, "failed to post message")
		}
	}

	user, icon, err := GetUserInfo(fromAPI, ev)
	fType, position, err := slack_lib.ConvertDisplayChannelName(fromAPI, ev)
	if err != nil {
		return errors.Wrap(err, "failed to convert channel name")
	}

	param := slack.PostMessageParameters{
		IconURL: icon,
	}
	username := user + "@" + strings.ToLower(fType[:1]) + ":" + position
	param.Username = username

	attachments := ev.Attachments

	// convert user id to user name in message
	msg, err = ConvertUserIdtoName(msg, ev, fromAPI)
	if err != nil {
		return errors.Wrap(err, "failed to convert id to name")
	}

	workspace := strings.TrimPrefix(aggrChannelName, config.PrefixSlackChannel)
	if msg != "" {
		respChannel, respTimestamp, err := toAPI.PostMessage(aggrChannelName, slack.MsgOptionText(msg, false), slack.MsgOptionPostMessageParameters(param))
		if err != nil {
			return errors.Wrap(err, "failed to post message")
		}
		store.SetSlackLog(workspace, ev.Timestamp, position, msg, respChannel, respTimestamp)
	}
	// if msg is blank, maybe bot_message (for example, twitter integration).
	// so, must post blank msg if this post have attachments.
	if attachments != nil {
		for _, attachment := range attachments {
			respChannel, respTimestamp, err := toAPI.PostMessage(aggrChannelName, slack.MsgOptionPostMessageParameters(param), slack.MsgOptionAttachments(attachment))
			if err != nil {
				return errors.Wrap(err, "failed to post message")
			}
			store.SetSlackLog(workspace, ev.Timestamp, position, msg, respChannel, respTimestamp)
		}
	}

	return nil
}
