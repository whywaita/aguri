package reply

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackutilsx"
	"github.com/whywaita/aguri/pkg/config"
	"github.com/whywaita/aguri/pkg/store"
	"github.com/whywaita/aguri/pkg/utils"
)

var (
	reChannel = regexp.MustCompile(`(\S+)@(\S+):(\S+)`)
)

func validateMessage(fromType slackutilsx.ChannelType, aggrChannelName string, ev *slack.MessageEvent) bool {
	if !strings.Contains(aggrChannelName, config.PrefixSlackChannel) {
		// not aggr channel
		return false
	}

	if ev.Msg.User == "USLACKBOT" {
		return false
	}

	if ev.Msg.Text == "" {
		// not normal message
		return false
	}

	if fromType != slackutilsx.CTypeChannel && fromType != slackutilsx.CTypeGroup {
		// TODO: implement other type
		return false
	}

	return true
}

func validateParsedMessage(userNames [][]string) bool {
	if len(userNames) == 0 {
		return false
	}

	return true
}

// HandleReplyMessage handle reply message from aggregated channel
func HandleReplyMessage(ctx context.Context, loggerMap *store.SyncLoggerMap) error {
	toAPI := store.GetConfigToAPI()
	rtm := toAPI.NewRTM(slack.RTMOptionUseStart(false))
	go rtm.ManageConnection()

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			if err := handleIncomingEvents(ctx, msg, toAPI, loggerMap); err != nil {
				log.Println(err)
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func handleIncomingEvents(ctx context.Context, msg slack.RTMEvent, toAPI *slack.Client, loggerMap *store.SyncLoggerMap) error {
	switch ev := msg.Data.(type) {
	case *slack.MessageEvent:
		fromType, aggrChName, err := utils.ConvertDisplayChannelNameMessageEvent(ctx, toAPI, ev)
		if err != nil {
			return fmt.Errorf("failed to convert display channel name: %w", err)

		}
		if !validateMessage(fromType, aggrChName, ev) {
			// invalid message
			return nil
		}

		workspace := strings.TrimPrefix(aggrChName, config.PrefixSlackChannel)
		if ev.ThreadTimestamp == "" {
			// maybe not in thread
			if err := handleReplyNotInThreadMessage(ctx, ev, workspace, loggerMap); err != nil {
				return fmt.Errorf("failed to handle receive message: %w", err)
			}
		}

		if err := handleReplyInThreadMessage(ctx, ev, workspace); err != nil {
			return fmt.Errorf("failed to handle reply message: %w", err)
		}

	case *slack.RTMError:
		return fmt.Errorf("detect rtm error: %s", ev.Error())
	}

	return nil
}

func handleReplyInThreadMessage(ctx context.Context, ev *slack.MessageEvent, workspace string) error {
	// reply message toSlack to fromSlack
	logData, err := store.GetSlackLog(workspace, ev.ThreadTimestamp)
	if err != nil {
		return fmt.Errorf("failed to get stored slack log: %w", err)
	}

	// Post
	api := store.GetSlackAPIInstance(workspace)
	param := slack.PostMessageParameters{
		AsUser: true,
	}

	_, _, err = api.PostMessageContext(ctx,
		logData.Channel,
		slack.MsgOptionText(ev.Text, false),
		slack.MsgOptionPostMessageParameters(param),
	)
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}

	return nil
}

func handleReplyNotInThreadMessage(ctx context.Context, ev *slack.MessageEvent, workspace string, loggerMap *store.SyncLoggerMap) error {
	logger, err := loggerMap.Load(workspace)
	if err != nil {
		return fmt.Errorf("failed to load loggerMap: %w", err)
	}

	if ev.User != "" {
		// write on toSlack
		if strings.HasPrefix(ev.Text, AguriCommandPrefix) {
			err := HandleAguriCommands(ctx, ev.Text, workspace)
			if err != nil {
				logger.Warn(err)
				return nil
			}
		}
	} else {
		// write on fromSlack
		err := saveSlackLogs(ev, workspace)
		if err != nil {
			logger.Warn(err)
			return nil
		}
	}

	return nil
}

func saveSlackLogs(ev *slack.MessageEvent, workspace string) error {
	// save slack log to store package
	if ev.Username == "" {
		return fmt.Errorf("Username is not found")
	}

	// parse username
	userNames := reChannel.FindAllStringSubmatch(ev.Username, -1)
	if !validateParsedMessage(userNames) {
		// miss regexp
		// or not channel
		return fmt.Errorf("failed to validate message")
	}

	if len(userNames[0]) < 3 {
		return fmt.Errorf("failed to get source channel name: %s", userNames[0])
	}
	chName := userNames[0][3]
	store.SetSlackLog(workspace, ev.Timestamp, chName, ev.Text, "", "")

	return nil
}
