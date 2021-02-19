package utils

import (
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/slack-go/slack"
)

func ConvertUnixToTime(inputTS string) (string, error) {
	// convert unixtime to Time
	times := strings.Split(inputTS, ".")
	unixtime := times[0]
	intTS, err := strconv.ParseInt(unixtime, 10, 64)
	if err != nil {
		return "", err
	}

	ts := time.Unix(intTS, 0).String()

	return ts, nil
}

func ConvertReadableName(api *slack.Client, ev *slack.MessageEvent) (slack.Msg, error) {
	var err error

	result := slack.Msg{}
	msg := ev.Msg

	copier.Copy(&result, &msg)

	// convert ID to Name
	rUser, err := api.GetUserInfo(msg.User)
	if err != nil {
		return slack.Msg{}, err
	}

	_, channelName, err := ConvertDisplayChannelName(api, ev)
	if err != nil {
		return slack.Msg{}, err
	}

	rTeam, err := api.GetTeamInfo()
	if err != nil {
		return slack.Msg{}, err
	}

	if err != nil {
		return slack.Msg{}, err
	}

	// convert time
	ts, err := ConvertUnixToTime(msg.Timestamp)
	if err != nil {
		return slack.Msg{}, err
	}

	result.User = rUser.Name
	result.Channel = channelName
	result.Team = rTeam.Name
	result.Timestamp = ts

	return result, nil
}
