package utils

import (
	"log"
	"regexp"
	"strings"

	"github.com/nlopes/slack"
	"github.com/whywaita/slack_lib"
)

var (
	reUser = regexp.MustCompile(`<@U(\S+)>`)
)

func checkExistChannel(api *slack.Client, searchName string) (bool, error) {
	// channel is exist => True
	channels, err := api.GetChannels(false)
	if err != nil {
		log.Println("[ERROR] checkExistChannel is fail")
		return false, err
	}

	for _, channel := range channels {
		if channel.Name == searchName {
			// if channel is exist, return true
			return true, nil
		}
	}

	return false, nil
}

func makeNewChannel(api *slack.Client, name string) error {
	var err error
	_, err = api.CreateChannel(name)
	if err != nil {
		log.Println("[ERROR] makeNewChannel is fail")
		log.Println(err)
		return err
	}

	return nil
}

func PostMessageToChannel(toAPI, fromAPI *slack.Client, ev *slack.MessageEvent, info *slack.Info, postChannelName string) (string, error) {
	// post aggregate message
	var err error

	isExist, err := checkExistChannel(toAPI, postChannelName)
	if err != nil {
		log.Println("[ERROR] postMessageToChannel is fail")
		return "", err
	}

	if (isExist == false) && (err == nil) {
		// if channel is not exist, make channel
		err = makeNewChannel(toAPI, postChannelName)
		if err != nil {
			log.Println("[ERROR] postMessageToChannel is fail")
			return "", err
		}
	}

	// get source username and channel, im, group
	user, usertype, err := slack_lib.ConvertDisplayUserName(fromAPI, ev, "")
	if err != nil {
		return "", err
	}

	fType, position, err := slack_lib.ConvertDisplayChannelName(fromAPI, ev)
	if err != nil {
		return "", err
	}

	icon := ""
	if usertype == "user" {
		u, err := fromAPI.GetUserInfo(ev.Msg.User)
		if err != nil {
			return "", err
		}
		icon = u.Profile.Image192
	}

	// convert user id to user name in message
	msg := ev.Text
	userIds := reUser.FindAllStringSubmatch(ev.Text, -1)
	if len(userIds) != 0 {
		for _, ids := range userIds {
			id := strings.TrimPrefix(ids[0], "<@")
			id = strings.TrimSuffix(id, ">")
			name, _, err := slack_lib.ConvertDisplayUserName(fromAPI, ev, id)
			if err != nil {
				log.Println(err)
				break
			}
			msg = strings.Replace(msg, id, name, -1)
		}
	}

	// post message
	param := slack.PostMessageParameters{
		IconURL: icon,
	}
	param.Username = user + "@" + strings.ToLower(fType[:1]) + ":" + position

	attachments := ev.Attachments

	if msg != "" {
		_, _, err = toAPI.PostMessage(postChannelName, slack.MsgOptionText(msg, false), slack.MsgOptionPostMessageParameters(param))
		if err != nil {
			log.Println("[ERROR] postMessageToChannel is fail")
			return "", err
		}
	}

	if attachments != nil {
		for _, attachment := range attachments {
			_, _, err = toAPI.PostMessage(postChannelName, slack.MsgOptionPostMessageParameters(param), slack.MsgOptionAttachments(attachment))
			if err != nil {
				log.Println("[ERROR] postMessageToChannel is fail")
				return "", err
			}
		}
	}

	return ev.Timestamp, nil
}
