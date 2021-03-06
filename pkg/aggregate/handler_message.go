package aggregate

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/whywaita/aguri/pkg/config"
	"github.com/whywaita/aguri/pkg/store"
	"github.com/whywaita/aguri/pkg/utils"
)

var (
	ErrAttachmentNotFound = errors.New("Detect Link Expand, but Attachment is not found")
)

func HandleMessageEvent(ev *slack.MessageEvent, fromAPI *slack.Client, workspace, lastTimestamp string, logger *logrus.Logger) string {
	var err error

	if lastTimestamp != ev.Timestamp {
		// if lastTimestamp == ev.Timestamp, that message is same.
		toChannelName := config.GetToChannelName(workspace)

		switch ev.SubType {
		case "message_changed":
			switch {
			case len(ev.SubMessage.Attachments) == 0:
				err = HandleMessageEdited(ev, fromAPI, workspace, toChannelName)
				if err != nil {
					logger.Warn(err)
					break
				}

			case len(ev.SubMessage.Attachments) >= 1:
				// message_changed and Text is null = URL link expand
				err = HandleMessageLinkExpand(ev, fromAPI, workspace, logger)
				if err != nil && err != ErrAttachmentNotFound {
					logger.Warn(err)
					break
				}
			}

		case "message_deleted":
			err = HandleMessageDeleted(ev, fromAPI, workspace, toChannelName)
			if err != nil {
				logger.Warn(err)
			}
		default:
			err = utils.PostMessageToChannel(store.GetConfigToAPI(), fromAPI, ev, ev.Text, toChannelName)
			if err != nil {
				logger.Warn(err)
			}
		}

		return ev.Timestamp
	}

	return lastTimestamp
}

func HandleMessageDeleted(ev *slack.MessageEvent, fromAPI *slack.Client, workspace, toChannelName string) error {
	d, err := store.GetSlackLog(workspace, ev.DeletedTimestamp)
	if err != nil {
		return errors.Wrap(err, "failed to get slack log from memory")
	}

	msg := fmt.Sprintf("Original Text:\n%v", d.Body)

	err = utils.PostMessageToChannel(store.GetConfigToAPI(), fromAPI, ev, msg, toChannelName)
	if err != nil {
		return errors.Wrap(err, "failed to post message")
	}

	return nil
}

func HandleMessageEdited(ev *slack.MessageEvent, fromAPI *slack.Client, workspace, toChannelName string) error {
	d, err := store.GetSlackLog(workspace, ev.SubMessage.Timestamp)
	if err != nil {
		return errors.Wrap(err, "failed to get slack log from memory")
	}

	msg := fmt.Sprintf("Edited From:\n%v", d.Body)
	msg += "\n\nEdited To:\n" + ev.SubMessage.Text

	err = utils.PostMessageToChannel(store.GetConfigToAPI(), fromAPI, ev, msg, toChannelName)
	if err != nil {
		return errors.Wrap(err, "failed to post message")
	}

	store.SetSlackLog(workspace, ev.SubMessage.Timestamp, d.Channel, ev.SubMessage.Text, "", "")

	return nil
}

func HandleMessageLinkExpand(ev *slack.MessageEvent, fromAPI *slack.Client, workspace string, logger *logrus.Logger) error {
	d, err := store.GetSlackLog(workspace, ev.SubMessage.Timestamp)
	if err != nil {
		return errors.Wrap(err, "failed to get slack log from memory")
	}

	switch {
	case len(ev.SubMessage.Attachments) == 0:
		return ErrAttachmentNotFound
	}
	_, _, _, err = store.GetConfigToAPI().UpdateMessage(d.ToAPICID, d.ToAPITS,
		slack.MsgOptionText(d.Body, false),
		slack.MsgOptionUpdate(d.ToAPITS),
		slack.MsgOptionAttachments(ev.SubMessage.Attachments...),
	)
	return err
}
