package stats

import (
	"fmt"
	"strings"
)

type ConversationMessage struct {
	// The sender of the message.
	SenderName *string `json:"sender_name"`
	// The content of the message.
	Content *string `json:"content"`
	// The timestamp of the message.
	TimestampMS *int64 `json:"timestamp_ms"`
	// Reactions to the message (only if the message has reactions)
	Reactions *[]*Reaction `json:"reactions,omitempty"`
	// Share information (only if the message is a share)
	Share *ShareInformation `json:"share,omitempty"`
	// Videos information (only if the message contains videos)
	Videos *[]*Video `json:"videos,omitempty"`
	// Photos information (only if the message contains photos)
	Photos *[]*Photo `json:"photos,omitempty"`
	// Audio information (only if the message contains audio)
	Audios *[]*Audio `json:"audio_files,omitempty"`
}

func (c *ConversationMessage) IsTextMessage() bool {
	return c.Content != nil && c.Share == nil
}

type ConversationRetrievalConfig struct {
	// The directory where the inbox is located.
	InboxDirectory string `json:"inbox_directory"`
	// Whether to include all subdirectories in the search.
	All bool `json:"all"`
	// List of directories (can be paths or regex patterns) to include in the search.
	Filters []string `json:"filters"`
	// The words to be used for counting occurrences in the conversation.
	CountOccurences []string `json:"count_occurrences"`
}

func (c *ConversationRetrievalConfig) CheckConfig() error {
	c.InboxDirectory = strings.TrimSpace(c.InboxDirectory)
	if c.InboxDirectory == "" {
		return fmt.Errorf("inbox_directory is required")
	}
	if len(c.Filters) == 0 && !c.All {
		return fmt.Errorf("filters are required if all is false")
	}
	return nil
}

type Conversation struct {
	// List of participants in the conversation.
	Participants []Participant `json:"participants"`
	// List of messages in the conversation.
	Messages []ConversationMessage `json:"messages"`
}

type Participant struct {
	Name string `json:"name"`
}

func (c *Conversation) Merge(other *Conversation) {
	// Merge participants
	participantsMap := make(map[string]struct{})
	for _, p := range c.Participants {
		participantsMap[p.Name] = struct{}{}
	}
	for _, p := range other.Participants {
		participantsMap[p.Name] = struct{}{}
	}
	c.Participants = make([]Participant, 0, len(participantsMap))
	for p := range participantsMap {
		c.Participants = append(c.Participants, Participant{Name: p})
	}

	// Merge messages
	c.Messages = append(c.Messages, other.Messages...)
}

type ShareInformation struct {
	Link                 string `json:"link"`
	ShareText            string `json:"share_text"`
	OriginalContentOwner string `json:"original_content_owner"`
}

type Video struct {
	Uri               string `json:"uri"`
	CreationTimestamp int64  `json:"creation_timestamp"`
}

type Photo struct {
	Uri               string `json:"uri"`
	CreationTimestamp int64  `json:"creation_timestamp"`
}

type Audio struct {
	Uri               string `json:"uri"`
	CreationTimestamp int64  `json:"creation_timestamp"`
}

type Reaction struct {
	// The reaction text
	Reaction string `json:"reaction"`
	// The timestamp of the reaction
	Timestamp int64 `json:"timestamp"`
	// The sender of the reaction
	Actor string `json:"actor"`
}
