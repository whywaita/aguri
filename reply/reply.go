package reply

import (
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
	reChannel    = regexp.MustCompile(`(\S+)@(\S+):(\S+)`)
	apiInstances = map[string]*slack.Client{}
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

func getSlackApiInstance(workspaceName string) *slack.Client {
	api, ok := apiInstances[workspaceName]
	if ok == false {
		// not found
		api = slack.New(store.GetConfigFromAPI(workspaceName))
		apiInstances[workspaceName] = api
	}

	return api
}

func HandleReplyMessage() {
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
				if ev.Username == "" {
					break
				}

				// parse username
				userNames := reChannel.FindAllStringSubmatch(ev.Username, -1)
				if !validateParsedMessage(userNames) {
					// miss regexp
					// or not channel
					break
				}

				if len(userNames[0]) < 3 {
					log.Printf("can't get source channel name: %v", userNames[0])
					break
				}
				chName := userNames[0][3]
				store.SetSlackLog(workspace, ev.Timestamp, chName, ev.Text)

				break
			}

			logData, err := store.GetSlackLog(workspace, ev.ThreadTimestamp)
			if err != nil {
				log.Println(err)
				break
			}

			// Post
			api := getSlackApiInstance(workspace)
			param := slack.PostMessageParameters{
				AsUser: true,
			}

			_, _, err = api.PostMessage(logData.Channel, slack.MsgOptionText(ev.Text, false), slack.MsgOptionPostMessageParameters(param))
			if err != nil {
				log.Println(err)
				break
			}

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		default:
			// Ignore
		}

	}
}
