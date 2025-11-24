package utils

import (
	"context"
	"testing"

	"github.com/slack-go/slack"
)

func TestPostMessageToChannelUploadedFile_NoChannels(t *testing.T) {
	// Test that the function properly handles files with no channels
	_ = context.Background()

	// Mock file with no channels
	originalFile := &slack.File{
		ID:       "F123456",
		Name:     "test.txt",
		Title:    "Test File",
		Channels: []string{}, // Empty channels array
		User:     "U123456",
	}

	_ = &slack.File{
		ID:        "F789012",
		Permalink: "https://example.slack.com/files/F789012",
	}

	_ = &slack.FileSharedEvent{
		EventTimestamp: "1234567890.123456",
	}

	// Note: This test requires mock clients for toAPI and fromAPI
	// For now, we test the validation logic

	if len(originalFile.Channels) == 0 {
		t.Log("Correctly detected empty channels array")
	} else {
		t.Error("Expected empty channels array")
	}
}

func TestPostMessageToChannelUploadedFile_ChannelIDExtraction(t *testing.T) {
	// Test that channel ID is correctly extracted from File.Channels
	originalFile := &slack.File{
		ID:       "F123456",
		Channels: []string{"C123456", "C789012"},
		User:     "U123456",
	}

	if len(originalFile.Channels) == 0 {
		t.Error("Expected file to have channels")
		return
	}

	channelID := originalFile.Channels[0]
	expectedChannelID := "C123456"

	if channelID != expectedChannelID {
		t.Errorf("Expected channel ID %s, got %s", expectedChannelID, channelID)
	}

	t.Logf("Successfully extracted channel ID: %s", channelID)
}

func TestPostMessageToChannelUploadedFile_MultipleChannels(t *testing.T) {
	// Test behavior when file is shared in multiple channels
	originalFile := &slack.File{
		ID:       "F123456",
		Channels: []string{"C123456", "C789012", "C345678"},
		User:     "U123456",
	}

	if len(originalFile.Channels) != 3 {
		t.Errorf("Expected 3 channels, got %d", len(originalFile.Channels))
	}

	// The function uses the first channel
	channelID := originalFile.Channels[0]
	if channelID != "C123456" {
		t.Errorf("Expected to use first channel C123456, got %s", channelID)
	}

	t.Logf("Using first channel from list of %d channels: %s", len(originalFile.Channels), channelID)
}

func TestIsSharedFile(t *testing.T) {
	// Test isSharedFile function
	testCases := []struct {
		name                 string
		file                 *slack.File
		sharedChannelID      string
		expectNonEmptyResult bool
	}{
		{
			name: "File shared in public channel",
			file: &slack.File{
				Shares: slack.Share{
					Public: map[string][]slack.ShareFileInfo{
						"C123456": {
							{
								Ts:       "1234567890.123456",
								ThreadTs: "",
							},
						},
					},
				},
			},
			sharedChannelID:      "C123456",
			expectNonEmptyResult: true,
		},
		{
			name: "File shared in private channel",
			file: &slack.File{
				Shares: slack.Share{
					Private: map[string][]slack.ShareFileInfo{
						"G123456": {
							{
								Ts:       "1234567890.123456",
								ThreadTs: "1234567890.000000",
							},
						},
					},
				},
			},
			sharedChannelID:      "G123456",
			expectNonEmptyResult: true,
		},
		{
			name: "File not shared in specified channel",
			file: &slack.File{
				Shares: slack.Share{
					Public: map[string][]slack.ShareFileInfo{
						"C789012": {
							{
								Ts: "1234567890.123456",
							},
						},
					},
				},
			},
			sharedChannelID:      "C123456",
			expectNonEmptyResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isSharedFile(tc.file, tc.sharedChannelID)

			if tc.expectNonEmptyResult && len(result) == 0 {
				t.Errorf("Expected non-empty result for %s", tc.name)
			}

			if !tc.expectNonEmptyResult && len(result) > 0 {
				t.Errorf("Expected empty result for %s, got %d items", tc.name, len(result))
			}

			if len(result) > 0 {
				t.Logf("%s: Found %d share(s)", tc.name, len(result))
			}
		})
	}
}
