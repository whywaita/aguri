package store

import (
	"errors"
	"strings"
)

var (
	wtc = map[string]string{} // "workspace,timestamp" : channel
)

var (
	ErrSourceChannelNotFound = errors.New("source channel is not found")
)

func SetSourceChannelName(workspace, timestamp, channelName string) {
	// register post to kv
	k := strings.Join([]string{workspace, timestamp}, ",")

	// TODO: gc
	wtc[k] = channelName
}

func GetSourceChannelName(workspace, timestamp string) (string, error) {
	parent := strings.Join([]string{workspace, timestamp}, ",")
	val, ok := wtc[parent]
	if ok == false {
		// TODO: if can't get channel name, search old message using slack API
		return "", ErrSourceChannelNotFound
	}

	return val, nil
}
