package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/nlopes/slack"
)

const (
	PrefixSlackChannel = "aggr-"
	PostUserName       = "slack-aggregator"
)

type Config struct {
	To   To              `toml:"to"`
	From map[string]From `toml:"from"`
}

type To struct {
	Token string `toml:"token"`
}

type From struct {
	Token string `toml:"token"`
}

type Froms struct {
	Team  string
	Token string
}

func loadConfig(configPath string) (string, []Froms, error) {
	var tomlConfig Config

	var toToken string
	var err error
	froms := []Froms{}
	from := Froms{}

	// load comfig file
	_, err = toml.DecodeFile(configPath, &tomlConfig)
	if err != nil {
		log.Println("[ERROR] loadConfig is fail", err)
		return "", nil, err
	}

	toToken = tomlConfig.To.Token

	for name, data := range tomlConfig.From {
		from.Team = name
		from.Token = data.Token
		froms = append(froms, from)
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

func dripValueByEV(fromAPI *slack.Client, ev *slack.MessageEvent, info *slack.Info) (string, string) {
	by := ""
	position := ""

	// user or bot
	if ev.Msg.BotID != "" {
		// this is bot
		byInfo, _ := fromAPI.GetBotInfo(ev.Msg.BotID)
		by = "Bot :" + byInfo.Name
	} else {
		byInfo, _ := fromAPI.GetUserInfo(ev.Msg.User)
		if byInfo != nil {
			by = "User :" + byInfo.Name
		} else {
			by = ""
		}
	}

	// public channel or private channel or group
	// Public channel prefix : C
	// Private channel prefix : G
	// Direct message prefix : D
	for _, c := range ev.Channel {
		if string(c) == "C" {
			poInfo, _ := fromAPI.GetChannelInfo(ev.Channel)
			position = "Channel :" + poInfo.Name
		} else if string(c) == "G" {
			poInfo, _ := fromAPI.GetGroupInfo(ev.Channel)
			position = "Group :" + poInfo.Name
		} else if string(c) == "D" {
			poInfo, _ := fromAPI.GetUserInfo(ev.Msg.User)
			position = "DM :" + poInfo.Name
		} else {
			position = " "
		}

		break
	}

	return by, position
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

	by, position := dripValueByEV(fromAPI, ev, info)
	param := slack.PostMessageParameters{}
	channelField := slack.AttachmentField{
		Title: "place",
		Value: position,
		Short: true,
	}
	userField := slack.AttachmentField{
		Title: "By",
		Value: by,
		Short: true,
	}
	attachment := slack.Attachment{
		Pretext: ev.Text,
	}
	attachment.Fields = []slack.AttachmentField{channelField, userField}
	param.Attachments = []slack.Attachment{attachment}
	param.Username = PostUserName
	_, _, err = toAPI.PostMessage(postChannelName, "", param)
	if err != nil {
		log.Println("[ERROR] postMessageToChannel is fail")
		return "", err
	}

	return ev.Timestamp, nil
}

func main() {
	var err error
	var wg sync.WaitGroup
	var lastTimestamp string
	var info *slack.Info

	// parse args
	var configPath = flag.String("config", "config.toml", "config file path")
	flag.Parse()

	// initialize
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)

	toToken, froms, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalln("[ERROR] ", err)
	}

	toAPI := slack.New(toToken)

	for _, from := range froms {
		wg.Add(1)
		// pass goroutine miss ref: http://qiita.com/sudix/items/67d4cad08fe88dcb9a6d
		fromToken := from.Token
		fromTeam := from.Team
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
					fmt.Printf("Message: %v\n", ev)
					if lastTimestamp != ev.Timestamp {
						lastTimestamp, err = postMessageToChannel(toAPI, fromAPI, ev, info, PrefixSlackChannel+fromTeam)
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
}
