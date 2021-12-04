package store

import (
	"fmt"
	"strings"
)

var (
	log = map[string]LogData{} // key: "workspace,timestamp"
)

// LogData is format of logging
type LogData struct {
	Channel        string
	Body           string
	ToAPIChannelID string
	ToAPITimestamp string
}

var (
	// ErrSourceChannelNotFound is error message for source channel is not found
	ErrSourceChannelNotFound = fmt.Errorf("source channel is not found")
)

// SetSlackLog set logging to memory
func SetSlackLog(workspace, timestamp, channelName, text, toAPIChannelID, toAPITimestamp string) {
	// register post to kv
	k := strings.Join([]string{workspace, timestamp}, ",")

	// TODO: gc
	log[k] = LogData{
		Channel:        channelName,
		Body:           text,
		ToAPIChannelID: toAPIChannelID,
		ToAPITimestamp: toAPITimestamp,
	}
}

// GetSlackLog get logging from memory
func GetSlackLog(workspace, timestamp string) (*LogData, error) {
	parent := strings.Join([]string{workspace, timestamp}, ",")
	val, ok := log[parent]
	if ok == false {
		// TODO: if can't get channel name, search old message using slack API
		return nil, ErrSourceChannelNotFound
	}

	return &val, nil
}
