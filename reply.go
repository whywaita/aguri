package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/nlopes/slack"
	"github.com/whywaita/slack_lib"
)

func validateMessage(fromType, aggrChannelName string, ev *slack.MessageEvent) bool {
	if !strings.Contains(aggrChannelName, PrefixSlackChannel) {
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

	if fromType != "channel" {
		// TODO: implement other type
		return false
	}

	return true
}

func replyMessage(toAPI *slack.Client, froms map[string]string) {
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

			workspace := strings.TrimPrefix(aggrChName, PrefixSlackChannel)

			if ev.ThreadTimestamp == "" {
				// maybe not in thread

				// register post to kv
				k := strings.Join([]string{workspace, ev.Timestamp}, ",")

				if ev.Username == "" {
					break
				}

				// parse username
				userNames := reChannel.FindAllStringSubmatch(ev.Username, -1)
				if len(userNames) == 0 || userNames[0][2] != "c" {
					// miss regexp
					// or not channel
					break
				}

				chName := userNames[0][3]

				// TODO: gc
				wtc[k] = chName

				break
			}

			parent := strings.Join([]string{workspace, ev.ThreadTimestamp}, ",")
			sourceChannelName := wtc[parent] // channel name

			// TODO: if can't get channel name, search old message using slack API

			// TODO: reuse api instance
			api := slack.New(froms[workspace])
			param := slack.PostMessageParameters{
				AsUser: true,
			}

			_, _, err = api.PostMessage(sourceChannelName, slack.MsgOptionText(ev.Text, false), slack.MsgOptionPostMessageParameters(param))
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
