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
// or a folder of conversations such as the inbox folder of the export. A file met
// through several sources, such as a folder and one of its subfolders, is read once.
//
// A file given as a source must be a conversation. In folders, the files that are not
// conversations or cannot be read, and the subfolders that cannot be listed, are
// skipped, and logged to logger.
func Load(sources []string, logger *zap.Logger) ([]*Conversation, error) {
	var conversations []*Conversation

	seen := map[string]struct{}{}

	for _, source := range sources {
		read, err := loadSource(source, seen, logger)
		if err != nil {
			return nil, err
		}

		conversations = append(conversations, read...)
	}

	return conversations, nil
}

// loadSource reads the conversations of source that are not in seen, and adds their
// files to seen.
func loadSource(source string, seen map[string]struct{}, logger *zap.Logger) ([]*Conversation, error) {
	info, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("reading the source: %w", err)
	}

	if !info.IsDir() {
		if !firstTime(seen, source) {
			return nil, nil
		}

		conversation, readErr := ReadFile(source)
		if readErr != nil {
			return nil, readErr
		}

		return []*Conversation{conversation}, nil
	}

	return loadFolder(source, seen, logger)
}

// loadFolder reads the conversation files under dir that are not in seen, skipping the
// other files, and adds them to seen.
func loadFolder(dir string, seen map[string]struct{}, logger *zap.Logger) ([]*Conversation, error) {
	files, err := jsonFiles(dir, logger)
	if err != nil {
		return nil, err
	}

	var conversations []*Conversation

	for _, file := range files {
		if !firstTime(seen, file) {
			logger.Debug("skipping file, it was already read", zap.String("file", file))

			continue
		}

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

	return conversations, nil
}

// firstTime records path in seen, and reports whether it was not in it yet.
func firstTime(seen map[string]struct{}, path string) bool {
	key, err := filepath.Abs(path)
	if err != nil {
		key = filepath.Clean(path)
	}

	if _, ok := seen[key]; ok {
		return false
	}

	seen[key] = struct{}{}

	return true
}

// jsonFiles returns the .json files under dir, sorted for a reproducible order. The
// subfolders that cannot be listed are skipped, and logged to logger.
func jsonFiles(dir string, logger *zap.Logger) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if path == dir {
				return err
			}

			logger.Warn("skipping folder, it could not be listed", zap.Error(err))

			if entry != nil && entry.IsDir() {
				return fs.SkipDir
			}

			return nil
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
