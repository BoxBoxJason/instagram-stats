package stats

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// CollectConversationsInDirectory opens all json files in the given directory
// It merges all conversations into one.
// returns them as a single Conversation struct.
func CollectConversationsInDirectory(directoryPath string) (*Conversation, error) {
	zap.L().Debug("Reading from inbox subdirectory", zap.String("directory", directoryPath))
	conversation := &Conversation{
		Participants: []Participant{},
		Messages:     []ConversationMessage{},
	}
	// Iterate over each file in the subdirectory
	subDir, err := os.Open(directoryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open subdirectory: %w", err)
	}
	defer subDir.Close()
	subChildren, err := subDir.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to read subdirectory: %w", err)
	}

	for _, subChild := range subChildren {
		if !subChild.IsDir() && strings.HasSuffix(subChild.Name(), ".json") {
			filePath := filepath.Join(directoryPath, subChild.Name())
			zap.L().Debug("Processing JSON conversation file", zap.String("file", filePath))
			partialConversation, err := ReadConversation(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read conversation from file %s: %w", filePath, err)
			}
			// Merge the conversation into the main conversation
			conversation.Merge(partialConversation)
		}
	}
	return conversation, nil
}

// ReadConversation reads a single JSON file and returns it as a Conversation struct.
func ReadConversation(filePath string) (*Conversation, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Decode the JSON into a Conversation struct
	var conversation Conversation
	if err := json.NewDecoder(file).Decode(&conversation); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return &conversation, nil
}
