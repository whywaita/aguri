package reply

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/nlopes/slack"
	"github.com/whywaita/aguri/config"
	"github.com/whywaita/aguri/store"
	"github.com/whywaita/slack_lib"
)

var (
	reChannel = regexp.MustCompile(`(\S+)@(\S+):(\S+)`)
)

func validateMessage(fromType, aggrChannelName string, ev *slack.MessageEvent) bool {
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

	if fromType != "channel" && fromType != "group" {
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

func HandleReplyMessage(loggerMap *store.SyncLoggerMap) {
	toAPI := store.GetConfigToAPI()
	rtm := toAPI.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			fromType, aggrChName, err := slack_lib.ConvertDisplayChannelName(toAPI, ev)
			if err != nil {
				log.Println(err)
				break
			}
			if !validateMessage(fromType, aggrChName, ev) {
				// invalid message
				break
			}

			workspace := strings.TrimPrefix(aggrChName, config.PrefixSlackChannel)

			if ev.ThreadTimestamp == "" {
				// maybe not in thread
				HandleReplyNotInThreadMessage(ev, workspace, loggerMap)
				break
			}

			err = HandleReplyInThreadMessage(ev, workspace)
			if err != nil {
				log.Println(err)
			}

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		default:
			// Ignore
		}

	}
}

func HandleReplyInThreadMessage(ev *slack.MessageEvent, workspace string) error {
	// reply message toSlack to fromSlack
	logData, err := store.GetSlackLog(workspace, ev.ThreadTimestamp)
	if err != nil {
		return err
	}

	// Post
	api := store.GetSlackApiInstance(workspace)
	param := slack.PostMessageParameters{
		AsUser: true,
	}

	_, _, err = api.PostMessage(logData.Channel, slack.MsgOptionText(ev.Text, false), slack.MsgOptionPostMessageParameters(param))
	if err != nil {
		return err
	}

	return nil
}

func HandleReplyNotInThreadMessage(ev *slack.MessageEvent, workspace string, loggerMap *store.SyncLoggerMap) {
	logger, err := loggerMap.Load(workspace)
	if err != nil {
		log.Println(err)
		return
	}

	if ev.User != "" {
		// write on toSlack
		if strings.HasPrefix(ev.Text, AguriCommandPrefix) {
			err := HandleAguriCommands(ev.Text, workspace)
			if err != nil {
				logger.Warn(err)
			}
		}
	} else {
		// write on fromSlack
		err := saveSlackLogs(ev, workspace)
		if err != nil {
			logger.Warn(err)
		}
	}

	return
}

func saveSlackLogs(ev *slack.MessageEvent, workspace string) error {
	// save slack log to store package
	if ev.Username == "" {
		return errors.New("Username is not found")
	}

	// parse username
	userNames := reChannel.FindAllStringSubmatch(ev.Username, -1)
	if !validateParsedMessage(userNames) {
		// miss regexp
		// or not channel
		return errors.New("failed to validate message")
	}

	if len(userNames[0]) < 3 {
		return fmt.Errorf("failed to get source channel name: %s", userNames[0])
	}
	chName := userNames[0][3]
	store.SetSlackLog(workspace, ev.Timestamp, chName, ev.Text, "", "")

	return nil
}
