package store

import (
	"strings"

	"github.com/pkg/errors"
)

var (
	log = map[string]LogData{} // key: "workspace,timestamp"
)

type LogData struct {
	Channel string
	Body    string
}

var (
	ErrSourceChannelNotFound = errors.New("source channel is not found")
)

func SetSlackLog(workspace, timestamp, channelName, text string) {
	// register post to kv
	k := strings.Join([]string{workspace, timestamp}, ",")

	// TODO: gc
	log[k] = LogData{
		Channel: channelName,
		Body:    text,
	}
}

func GetSlackLog(workspace, timestamp string) (*LogData, error) {
	parent := strings.Join([]string{workspace, timestamp}, ",")
	val, ok := log[parent]
	if ok == false {
		// TODO: if can't get channel name, search old message using slack API
		return nil, errors.Wrap(ErrSourceChannelNotFound, "")
	}

	return &val, nil
}
