package stats

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

type ActivityType int

const (
	// Activity types for messages
	ACTIVITY_UNKNOWN ActivityType = iota
	ACTIVITY_MESSAGE
	ACTIVITY_MEDIA
	ACTIVITY_AUDIO
	ACTIVITY_REEL
	ACTIVITY_REACTION
	CONVERSATION_FILE = "conversation.json"
	USER_FILE         = "%s.json"
)

type StatsRetrievalConfig struct {
	WordsToSearch *map[string]struct{}
}

type OutputConfig struct {
	// Directory where the output files will be saved.
	Directory string `json:"directory"`
	// Whether raw json file should be hidden
	NoJson bool `json:"no_json"`
}

// NewStatsRetrievalConfig initializes a new StatsRetrievalConfig with the provided words.
// If words is nil, it initializes an empty map.
func NewStatsRetrievalConfig(words *[]string) *StatsRetrievalConfig {
	wordsToSearch := make(map[string]struct{})
	if words != nil {
		for _, word := range *words {
			wordsToSearch[word] = struct{}{}
		}
	}
	return &StatsRetrievalConfig{
		WordsToSearch: &wordsToSearch,
	}
}

type ProcessingInformation struct {
	Users                 map[string]*ProcessedUser `json:"users"`
	ProcessedConversation *ProcessedConversation    `json:"processed_conversation"`
}

// GetUser retrieves a user from the processing information.
// If the user does not exist, it creates a new user with default values.
// It also initializes the user's heatmap with 7 days and 24 hours.
func (p *ProcessingInformation) GetUser(name string) *ProcessedUser {
	if user, ok := p.Users[name]; ok && user != nil {
		return user
	} else {
		// Create a new user if it doesn't exist
		user := &ProcessedUser{
			Name: name,
			// Initialize the user with default values
			ActivityStats: *NewActivityStats(),
		}
		p.Users[name] = user
		return user
	}
}

// Save outputs the processing information to a file in the specified output directory.
// There is a file for each user and one for the conversation.
func (p *ProcessingInformation) Save(outputDir string) error {
	zap.L().Info("Saving processing information", zap.String("save_dir", outputDir))
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save the conversation
	convoFile := filepath.Join(outputDir, CONVERSATION_FILE)
	convoJson, err := json.Marshal(p.ProcessedConversation)
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}
	if err := os.WriteFile(convoFile, convoJson, 0644); err != nil {
		return fmt.Errorf("failed to write conversation file: %w", err)
	} else {
		zap.L().Debug("Saved conversation file", zap.String("file", convoFile))
	}

	// Save each user
	for _, user := range p.Users {
		userFile := filepath.Join(outputDir, fmt.Sprintf(USER_FILE, user.Name))
		userJson, err := json.Marshal(user)
		if err != nil {
			return fmt.Errorf("failed to marshal user %s: %w", user.Name, err)
		}
		if err := os.WriteFile(userFile, userJson, 0644); err != nil {
			return fmt.Errorf("failed to write user file %s: %w", userFile, err)
		} else {
			zap.L().Debug("Saved user file", zap.String("file", userFile))
		}
	}
	return nil
}

type ActivityStats struct {
	// Number of (text) messages sent by the user.
	MessagesCount int `json:"messages_count"`
	// Number of (media: video, image) messages sent by the user.
	MediaMessagesCount int `json:"media_messages_count"`
	// Number of (audio) messages sent by the user.
	AudioMessagesCount int `json:"audio_messages_count"`
	// Number of reactions sent by the user.
	ReactionsCount int `json:"reactions_count"`
	// Number of reels sent by the user.
	ReelsCount int `json:"reels_count"`
	// Reactions received by the user.
	ReactionsReceivedCount int             `json:"reactions_received_count"`
	Heatmap                *[][]int        `json:"heatmap"`
	MatchFrequency         *map[string]int `json:"match_frequency"`
	Vocabulary             *map[string]int `json:"vocabulary"`
	VocabularyCount        int             `json:"distinct_words"`
	DailyActivity          *map[string]int `json:"daily_activity"`
	WeeklyActivity         *map[string]int `json:"weekly_activity"`
	MonthlyActivity        *map[string]int `json:"monthly_activity"`
	YearlyActivity         *map[string]int `json:"yearly_activity"`
	StartDate              int64           `json:"start_date"`
	EndDate                int64           `json:"end_date"`
}

func (a *ActivityStats) AddActivity(timestamp int64, activityType ActivityType) {
	t := time.Unix(timestamp, 0)
	year, month, day := t.Date()
	_, week := t.ISOWeek()
	dayOfWeek := int(t.Weekday())
	hour := t.Hour()

	a.UpdateDailyActivity(fmt.Sprintf("%d-%02d-%02d", year, month, day), activityType)
	a.UpdateWeeklyActivity(fmt.Sprintf("%d-%02d", year, week), activityType)
	a.UpdateMonthlyActivity(fmt.Sprintf("%d-%02d", year, month), activityType)
	a.UpdateYearlyActivity(fmt.Sprintf("%d", year), activityType)

	if a.Heatmap != nil {
		(*a.Heatmap)[dayOfWeek][hour]++
	}

	a.UpdateTimestamp(timestamp)
}

func (a *ActivityStats) UpdateDailyActivity(day string, activityType ActivityType) {
	if a.DailyActivity != nil {
		(*a.DailyActivity)[day]++
	}
}
func (a *ActivityStats) UpdateWeeklyActivity(week string, activityType ActivityType) {
	if a.WeeklyActivity != nil {
		(*a.WeeklyActivity)[week]++
	}
}
func (a *ActivityStats) UpdateMonthlyActivity(month string, activityType ActivityType) {
	if a.MonthlyActivity != nil {
		(*a.MonthlyActivity)[month]++
	}
}
func (a *ActivityStats) UpdateYearlyActivity(year string, activityType ActivityType) {
	if a.YearlyActivity != nil {
		(*a.YearlyActivity)[year]++
	}
}

func (a *ActivityStats) UpdateTimestamp(timestamp int64) {
	if timestamp < a.StartDate {
		a.StartDate = timestamp
	}
	if timestamp > a.EndDate {
		a.EndDate = timestamp
	}
}

type ProcessedUser struct {
	ActivityStats
	// Name of the user.
	Name string `json:"name"`
}

type ProcessedConversation struct {
	ActivityStats
	// List of participants in the conversation.
	Participants map[string]struct{} `json:"participants"`
}

func NewActivityStats() *ActivityStats {
	// Initialize the heatmap with 7 days and 24 hours
	heatmap := make([][]int, 7)
	for i := range heatmap {
		heatmap[i] = make([]int, 24)
	}
	return &ActivityStats{
		MessagesCount:          0,
		MediaMessagesCount:     0,
		AudioMessagesCount:     0,
		ReactionsCount:         0,
		ReelsCount:             0,
		ReactionsReceivedCount: 0,
		Vocabulary:             &map[string]int{},
		VocabularyCount:        0,
		MatchFrequency:         &map[string]int{},
		Heatmap:                &heatmap,
		DailyActivity:          &map[string]int{},
		WeeklyActivity:         &map[string]int{},
		MonthlyActivity:        &map[string]int{},
		YearlyActivity:         &map[string]int{},
		StartDate:              0,
		EndDate:                0,
	}
}

func NewProcessedConversation() *ProcessedConversation {
	return &ProcessedConversation{
		Participants:  make(map[string]struct{}),
		ActivityStats: *NewActivityStats(),
	}
}

func NewProcessingInformation() *ProcessingInformation {
	return &ProcessingInformation{
		Users:                 make(map[string]*ProcessedUser),
		ProcessedConversation: NewProcessedConversation(),
	}
}
