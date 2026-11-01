package export

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.uber.org/zap"
)

// ErrNotConversation is returned for a JSON file that is not a conversation, such as
// the secret_conversations.json file of the export.
var ErrNotConversation = errors.New("not an Instagram conversation file")

// ReadFile reads the conversation file at path.
func ReadFile(path string) (*Conversation, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var conversation Conversation

	err = json.Unmarshal(data, &conversation)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	if conversation.Participants == nil && conversation.Messages == nil {
		return nil, fmt.Errorf("%s: %w", path, ErrNotConversation)
	}

	return &conversation, nil
}

// Load reads the conversations of sources. A source is a conversation file, or a
// folder whose .json files are read, in its subfolders too: a conversation folder,
// or a folder of conversations such as the inbox folder of the export.
//
// A file given as a source must be a conversation. In folders, the files that are not
// conversations or cannot be read are skipped, and logged to logger.
func Load(sources []string, logger *zap.Logger) ([]*Conversation, error) {
	var conversations []*Conversation

	for _, source := range sources {
		info, err := os.Stat(source)
		if err != nil {
			return nil, fmt.Errorf("reading the source: %w", err)
		}

		if !info.IsDir() {
			conversation, readErr := ReadFile(source)
			if readErr != nil {
				return nil, readErr
			}

			conversations = append(conversations, conversation)

			continue
		}

		files, err := jsonFiles(source)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			logger.Debug("reading conversation file", zap.String("file", file))

			conversation, readErr := ReadFile(file)

			switch {
			case errors.Is(readErr, ErrNotConversation):
				logger.Debug("skipping file, it is not a conversation", zap.String("file", file))
			case readErr != nil:
				logger.Warn("skipping file, it could not be read", zap.Error(readErr))
			default:
				conversations = append(conversations, conversation)
			}
		}
	}

	return conversations, nil
}

// jsonFiles returns the .json files under dir, sorted for a reproducible order.
func jsonFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() && strings.EqualFold(filepath.Ext(path), ".json") {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing the files of %s: %w", dir, err)
	}

	sort.Strings(files)

	return files, nil
}
