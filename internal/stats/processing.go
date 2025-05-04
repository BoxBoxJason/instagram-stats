package stats

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"go.uber.org/zap"
)

// ProcessConversations opens all json files in the given directory
func ProcessConversations(collectionConfig *ConversationRetrievalConfig, outputConfig *OutputConfig) error {
	zap.L().Debug("Reading from inbox directory", zap.String("directory", collectionConfig.InboxDirectory))
	dirEntries, err := os.ReadDir(collectionConfig.InboxDirectory)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	var filterRegex *regexp.Regexp
	if !collectionConfig.All && len(collectionConfig.Filters) > 0 {
		filterRegex = regexp.MustCompile(strings.Join(collectionConfig.Filters, "|"))
	}

	// Pre-filter valid subdirectories
	var validDirs []string
	for _, entry := range dirEntries {
		if entry.IsDir() && (collectionConfig.All || filterRegex.MatchString(entry.Name())) {
			validDirs = append(validDirs, entry.Name())
		}
	}

	zap.L().Info("Found valid subdirectories", zap.Int("count", len(validDirs)))

	// Iterate over the conversations in the subdirectories
	for i, dirName := range validDirs {
		zap.L().Debug("Processing conversation", zap.String("conversation", dirName), zap.Int("index", i+1), zap.Int("total", len(validDirs)))
		childPath := filepath.Join(collectionConfig.InboxDirectory, dirName)
		conversation, err := CollectConversationsInDirectory(childPath)
		if err != nil {
			return fmt.Errorf("failed to collect conversations in directory %s: %w", childPath, err)
		}

		if len(conversation.Messages) > 0 {
			processingInformation := conversation.ProcessConversation(NewStatsRetrievalConfig(&collectionConfig.CountOccurences))
			processingInformation.Save(filepath.Join(outputConfig.Directory, dirName))
		}
		zap.L().Info("Finished processing conversation", zap.String("conversation", dirName), zap.String("remaining", fmt.Sprintf("%d/%d", len(validDirs)-i-1, len(validDirs))))
	}

	return nil
}

// ProcessConversation parses each message in the conversation and creates all statistics
// for each user in the conversation.
func (c *Conversation) ProcessConversation(config *StatsRetrievalConfig) *ProcessingInformation {
	// Process the conversation messages in a single loop
	processingInfo := NewProcessingInformation()
	for _, message := range c.Messages {
		message.ProcessMessage(config, processingInfo)
	}
	return processingInfo
}

// ProcessMessage processes a single message in the conversation.
//
// Responsible for individual message processing and counting.
func (m *ConversationMessage) ProcessMessage(config *StatsRetrievalConfig, processingInfo *ProcessingInformation) {
	sender := processingInfo.GetUser(*m.SenderName)
	m.CountMessage(processingInfo.ProcessedConversation, sender, config.WordsToSearch)
}

// CountMessage is used to increment the corresponding counters for the sender
// based on the type of message sent.
func (m *ConversationMessage) CountMessage(conversation *ProcessedConversation, sender *ProcessedUser, wordsToSearch *map[string]struct{}) {
	activity := m.DetermineActivityType()

	switch activity {
	case ACTIVITY_MESSAGE:
		sender.MessagesCount++
		m.CountVocabulary(conversation, sender, wordsToSearch)
		conversation.MessagesCount++
	case ACTIVITY_MEDIA:
		if m.Videos != nil {
			sender.MediaMessagesCount += len(*m.Videos)
			conversation.MediaMessagesCount += len(*m.Videos)
		}
		if m.Photos != nil {
			sender.MediaMessagesCount += len(*m.Photos)
			conversation.MediaMessagesCount += len(*m.Photos)
		}

	case ACTIVITY_AUDIO:
		sender.AudioMessagesCount += len(*m.Audios)
		conversation.AudioMessagesCount += len(*m.Audios)

	case ACTIVITY_REEL:
		sender.ReelsCount++
		conversation.ReelsCount++
	}

	if activity != ACTIVITY_UNKNOWN {
		sender.AddActivity(*m.TimestampMS/1000, activity)
		conversation.AddActivity(*m.TimestampMS/1000, activity)
	}
}

// DetermineActivityType determines the type of activity based on the message content.
// It checks if the message is a text message, media message, audio message, or a reel.
func (m *ConversationMessage) DetermineActivityType() ActivityType {
	if m.IsTextMessage() {
		return ACTIVITY_MESSAGE
	} else if m.Videos != nil || m.Photos != nil {
		return ACTIVITY_MEDIA
	} else if m.Audios != nil {
		return ACTIVITY_AUDIO
	} else if m.Share != nil {
		return ACTIVITY_REEL
	}
	return ACTIVITY_UNKNOWN
}

// CountVocabulary is used to determine the number of separate words that appear in a user's messages.
// It also counts the number of times each word appears in the user's messages.
//
// Additionally, it counts the number of matches of keywords if they are provided
func (m *ConversationMessage) CountVocabulary(conversation *ProcessedConversation, sender *ProcessedUser, wordsToSearch *map[string]struct{}) {
	searchWordsPresent := wordsToSearch != nil && len(*wordsToSearch) > 0
	if m.Content == nil {
		return
	}

	words := strings.Fields(*m.Content)
	for _, rawWord := range words {
		word := normalizeWord(rawWord)
		if word == "" {
			continue
		}
		(*sender.Vocabulary)[word]++
		(*conversation.Vocabulary)[word]++
		if searchWordsPresent {
			if _, ok := (*wordsToSearch)[word]; ok {
				(*sender.MatchFrequency)[word]++
				(*conversation.MatchFrequency)[word]++
			}
		}
	}
}

// normalizeWord normalizes a word by converting it to lowercase and removing non-letter and non-number characters.
// It also trims leading and trailing whitespace.
func normalizeWord(word string) string {
	return strings.ToLower(strings.TrimFunc(word, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	}))
}

// ProcessReaction processes a single reaction in the conversation.
// Adds the reaction to the sender's activity and increments the conversation's reaction count.
func (r *Reaction) ProcessReaction(processingInfo *ProcessingInformation) {
	sender := processingInfo.GetUser(r.Actor)
	sender.ReactionsCount++
	sender.AddActivity(r.Timestamp, ACTIVITY_REACTION)

	processingInfo.ProcessedConversation.ReactionsCount++
	processingInfo.ProcessedConversation.AddActivity(r.Timestamp, ACTIVITY_REACTION)
}
