package aggregate

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/whywaita/aguri/pkg/config"
	"github.com/whywaita/aguri/pkg/store"
	"github.com/whywaita/aguri/pkg/utils"
)

var (
	// ErrAttachmentNotFound is error message for "Detect Link Expand, but Attachment is not found"
	ErrAttachmentNotFound = fmt.Errorf("Detect Link Expand, but Attachment is not found")
)

// HandleMessageEvent handle message event
func HandleMessageEvent(ctx context.Context, ev *slack.MessageEvent, fromAPI *slack.Client, workspace, lastTimestamp string, logger *logrus.Logger) string {
	var err error

	if lastTimestamp != ev.Timestamp {
		// if lastTimestamp == ev.Timestamp, that message is same.
		toChannelName := config.GetToChannelName(workspace)

		switch ev.SubType {
		case "message_changed":
			switch {
			case len(ev.SubMessage.Attachments) == 0:
				err = handleMessageEdited(ctx, ev, fromAPI, workspace, toChannelName)
				if err != nil {
					logger.Warn(err)
					break
				}

			case len(ev.SubMessage.Attachments) >= 1:
				// message_changed and Text is null = URL link expand
				if err = handleMessageLinkExpand(ctx, ev, fromAPI, workspace, logger); err != nil && err != ErrAttachmentNotFound {
					logger.Warn(err)
					break
				}
			}

		case "message_deleted":
			err = handleMessageDeleted(ctx, ev, fromAPI, workspace, toChannelName)
			if err != nil {
				logger.Warn(err)
			}
		default:
			err = utils.PostMessageToChannelMessageEvent(ctx, store.GetConfigToAPI(), fromAPI, ev, ev.Text, toChannelName)
			if err != nil {
				logger.Warn(err)
			}
		}

		return ev.Timestamp
	}

	return lastTimestamp
}

func handleMessageDeleted(ctx context.Context, ev *slack.MessageEvent, fromAPI *slack.Client, workspace, toChannelName string) error {
	d, err := store.GetSlackLog(workspace, ev.DeletedTimestamp)
	if err != nil {
		return fmt.Errorf("failed to get slack log from memory: %w", err)
	}

	msg := fmt.Sprintf("Original Text:\n%v", d.Body)

	err = utils.PostMessageToChannelMessageEvent(ctx, store.GetConfigToAPI(), fromAPI, ev, msg, toChannelName)
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}

	return nil
}

func handleMessageEdited(ctx context.Context, ev *slack.MessageEvent, fromAPI *slack.Client, workspace, toChannelName string) error {
	d, err := store.GetSlackLog(workspace, ev.SubMessage.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to get slack log from memory: %w", err)
	}

	msg := fmt.Sprintf("Edited From:\n%v", d.Body)
	msg += "\n\nEdited To:\n" + ev.SubMessage.Text

	err = utils.PostMessageToChannelMessageEvent(ctx, store.GetConfigToAPI(), fromAPI, ev, msg, toChannelName)
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}

	store.SetSlackLog(workspace, ev.SubMessage.Timestamp, d.Channel, ev.SubMessage.Text, "", "")

	return nil
}

func handleMessageLinkExpand(ctx context.Context, ev *slack.MessageEvent, fromAPI *slack.Client, workspace string, logger *logrus.Logger) error {
	d, err := store.GetSlackLog(workspace, ev.SubMessage.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to get slack log from memory: %w", err)
	}

	switch {
	case len(ev.SubMessage.Attachments) == 0:
		return ErrAttachmentNotFound
	}
	if _, _, _, err = store.GetConfigToAPI().UpdateMessageContext(ctx,
		d.ToAPIChannelID, d.ToAPITimestamp,
		slack.MsgOptionText(d.Body, false),
		slack.MsgOptionUpdate(d.ToAPITimestamp),
		slack.MsgOptionAttachments(ev.SubMessage.Attachments...),
	); err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}
	return nil
}
