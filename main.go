package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/nlopes/slack"
	"github.com/whywaita/slack_lib"
)

const (
	PrefixSlackChannel = "aggr-"
)

var (
	reUser    = regexp.MustCompile("<@U.*>")
	reChannel = regexp.MustCompile("(.*)@(.*)")
	wtc       = map[string]string{} // "workspace,timestamp" : channel
)

type Config struct {
	To   To              `toml:"to"`
	From map[string]From `toml:"from"`
}

// for toml
type To struct {
	Token string `toml:"token"`
}

// for toml
type From struct {
	Token string `toml:"token"`
}

func loadConfig(configPath string) (string, map[string]string, error) {
	var tomlConfig Config

	var toToken string
	var err error
	froms := map[string]string{}

	// load comfig file
	_, err = toml.DecodeFile(configPath, &tomlConfig)
	if err != nil {
		log.Println("[ERROR] loadConfig is fail", err)
		return "", nil, err
	}

	toToken = tomlConfig.To.Token

	for name, data := range tomlConfig.From {
		froms[name] = data.Token
	}

	return toToken, froms, err
}

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

func postMessageToChannel(toAPI, fromAPI *slack.Client, ev *slack.MessageEvent, info *slack.Info, postChannelName string) (string, error) {
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

	_, position, err := slack_lib.ConvertDisplayChannelName(fromAPI, ev)
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
			fmt.Println(ids[0])
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
	attachment := slack.Attachment{
		Pretext: msg,
	}
	param.Attachments = []slack.Attachment{attachment}
	param.Username = user + "@" + position

	_, _, err = toAPI.PostMessage(postChannelName, slack.MsgOptionText(msg, false), slack.MsgOptionPostMessageParameters(param))
	if err != nil {
		log.Println("[ERROR] postMessageToChannel is fail")
		return "", err
	}

	return ev.Timestamp, nil
}

func catchMessage(froms map[string]string, toAPI *slack.Client) error {
	var wg sync.WaitGroup
	var info *slack.Info
	var lastTimestamp string
	var err error

	for team, token := range froms {
		wg.Add(1)
		// pass goroutine miss ref: http://qiita.com/sudix/items/67d4cad08fe88dcb9a6d
		fromToken := token
		fromTeam := team
		go func() {
			fromAPI := slack.New(fromToken)
			rtm := fromAPI.NewRTM()
			go rtm.ManageConnection()
			for msg := range rtm.IncomingEvents {
				switch ev := msg.Data.(type) {
				case *slack.HelloEvent:
					// Ignore Hello
				case *slack.ConnectedEvent:
					info = ev.Info
				case *slack.MessageEvent:
					// fmt.Printf("Message: %v\n", ev)

					if lastTimestamp != ev.Timestamp {
						chName := PrefixSlackChannel + strings.ToLower(fromTeam)

						lastTimestamp, err = postMessageToChannel(toAPI, fromAPI, ev, info, chName)
						if err != nil {
							log.Println(err)
						}
					}
				case *slack.RTMError:
					fmt.Printf("Error: %s\n", ev.Error())
				default:
					// Ignore
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	return nil
}

func replyMessage(toAPI *slack.Client, froms map[string]string) {
	rtm := toAPI.NewRTM()
	go rtm.ManageConnection()
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore Hello
			//case *slack.ConnectedEvent:
			//	info = ev.Info
		case *slack.MessageEvent:
			fromType, aggrChName, err := slack_lib.ConvertDisplayChannelName(toAPI, ev)
			if err != nil {
				log.Println(err)
				break
			}
			if !strings.Contains(aggrChName, PrefixSlackChannel) {
				// not aggr channel
				break
			}

			if fromType != "channel" {
				// TODO: implement other type
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
				chName := userNames[0][2]

				// TODO: gc
				wtc[k] = chName

				break
			}

			parent := strings.Join([]string{workspace, ev.ThreadTimestamp}, ",")
			sourceChannelName := wtc[parent] // channel name

			// TODO: if can't get channel name, search use slack API

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

func main() {
	// parse args
	var configPath = flag.String("config", "config.toml", "config file path")
	flag.Parse()

	// initialize
	//logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	//slack.SetLogger(logger)
	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)

	toToken, froms, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalln("[ERROR] ", err)
	}

	toAPI := slack.New(toToken)
	go replyMessage(toAPI, froms)

	err = catchMessage(froms, toAPI)
	if err != nil {
		log.Fatal(err)
	}
}
