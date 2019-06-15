package aggregate

import (
	"fmt"
	"log"
	"strings"

	"github.com/whywaita/aguri/store"

	"github.com/nlopes/slack"
	"github.com/whywaita/aguri/config"
	"github.com/whywaita/aguri/utils"
)

func HandleMessageEvent(ev *slack.MessageEvent, fromAPI *slack.Client, workspace, lastTimestamp string) string {
	var err error

	if lastTimestamp != ev.Timestamp {
		// if lastTimestamp == eve.Timestamp, that message is same.
		toChannelName := config.PrefixSlackChannel + strings.ToLower(workspace)

		switch ev.SubType {
		case "message_changed":
			err = HandleMessageEdited(ev, fromAPI, workspace, toChannelName)
		case "message_deleted":
			err = HandleMessageDeleted(ev, fromAPI, workspace, toChannelName)
			if err != nil {
				log.Println(err)
			}
		default:
			err = utils.PostMessageToChannel(store.GetConfigToAPI(), fromAPI, ev, ev.Text, toChannelName)
			if err != nil {
				log.Println(err)
			}
		}

		return ev.Timestamp
	}

	return lastTimestamp
}

func HandleMessageDeleted(ev *slack.MessageEvent, fromAPI *slack.Client, workspace, toChannelName string) error {
	d, err := store.GetSlackLog(workspace, ev.DeletedTimestamp)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("Original Text:\n%v", d.Body)

	err = utils.PostMessageToChannel(store.GetConfigToAPI(), fromAPI, ev, msg, toChannelName)
	if err != nil {
		return err
	}

	return nil
}

func HandleMessageEdited(ev *slack.MessageEvent, fromAPI *slack.Client, workspace, toChannelName string) error {
	d, err := store.GetSlackLog(workspace, ev.SubMessage.Timestamp)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("Original Text:\n%v", d.Body)
	msg += "\n\nEdited Text\n" + ev.SubMessage.Text

	err = utils.PostMessageToChannel(store.GetConfigToAPI(), fromAPI, ev, msg, toChannelName)
	if err != nil {
		return err
	}

	return nil
}
