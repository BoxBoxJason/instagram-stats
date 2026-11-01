// Package export reads the conversations of an Instagram data export, in JSON format.
//
// Each conversation of the export is a folder holding one or more message_<n>.json
// files, each one a [Conversation] holding a batch of its messages.
package export

import (
	"encoding/json"
	"time"
)

// Conversation is the content of one message_<n>.json file of the export.
type Conversation struct {
	IsPending          *bool             `json:"is_pending,omitempty"`
	JoinableMode       *JoinableMode     `json:"joinable_mode,omitempty"`
	Title              Text              `json:"title"`
	ThreadPath         string            `json:"thread_path"`
	Participants       []Participant     `json:"participants"`
	Messages           []Message         `json:"messages"`
	MagicWords         []json.RawMessage `json:"magic_words"`
	IsStillParticipant bool              `json:"is_still_participant"`
}

// Participant is a member of a conversation.
type Participant struct {
	// Name is the display name of the participant, not their username.
	Name Text `json:"name"`
}

// JoinableMode describes how a group conversation can be joined.
type JoinableMode struct {
	Link string `json:"link"`
	Mode int    `json:"mode"`
}

// Message is a message of a conversation. Besides its text, a message can hold a
// share (a reel, a post, a profile...), photos, videos, audio files or a call.
type Message struct {
	Share                             *Share     `json:"share,omitempty"`
	CallDuration                      *int64     `json:"call_duration,omitempty"`
	SenderName                        Text       `json:"sender_name"`
	Content                           Text       `json:"content,omitempty"`
	Photos                            []Media    `json:"photos,omitempty"`
	Videos                            []Media    `json:"videos,omitempty"`
	AudioFiles                        []Media    `json:"audio_files,omitempty"`
	Reactions                         []Reaction `json:"reactions,omitempty"`
	TimestampMs                       int64      `json:"timestamp_ms"`
	IsGeoblockedForViewer             bool       `json:"is_geoblocked_for_viewer"`
	IsUnsentImageByMessengerKidParent bool       `json:"is_unsent_image_by_messenger_kid_parent"`
}

// Time returns the time the message was sent at.
func (m Message) Time() time.Time {
	return time.UnixMilli(m.TimestampMs)
}

// Kind is what a message mainly holds.
type Kind int

const (
	// KindText is a plain message: text, or nothing the export kept (a sticker...).
	KindText Kind = iota
	// KindShare is a shared reel, post or profile.
	KindShare
	// KindMedia holds photos or videos.
	KindMedia
	// KindAudio is a voice message.
	KindAudio
	// KindCall logs an audio or video call.
	KindCall
)

// Kind returns what the message mainly holds. A message holding several kinds of
// content is classified by the first of share, photos or videos, audio files and call.
func (m Message) Kind() Kind {
	switch {
	case m.Share != nil:
		return KindShare
	case len(m.Photos) > 0 || len(m.Videos) > 0:
		return KindMedia
	case len(m.AudioFiles) > 0:
		return KindAudio
	case m.CallDuration != nil:
		return KindCall
	default:
		return KindText
	}
}

// Share is the content shared by a message.
type Share struct {
	// Link is the URL of the shared content.
	Link string `json:"link,omitempty"`
	// ShareText is the caption of the shared content.
	ShareText Text `json:"share_text,omitempty"`
	// OriginalContentOwner is the username of the author of the shared content.
	OriginalContentOwner Text `json:"original_content_owner,omitempty"`
	// ProfileShareName and ProfileShareUsername are set when a profile is shared.
	ProfileShareName     Text `json:"profile_share_name,omitempty"`
	ProfileShareUsername Text `json:"profile_share_username,omitempty"`
}

// Media is a photo, a video or an audio file sent in a message.
type Media struct {
	// URI is the path of the file in the export, relative to its root folder.
	URI string `json:"uri"`
	// CreationTimestamp is the Unix time the file was created at, in seconds. It is
	// missing for some files.
	CreationTimestamp int64 `json:"creation_timestamp,omitempty"`
}

// Reaction is a reaction a participant added to a message.
type Reaction struct {
	// Reaction is the emoji of the reaction.
	Reaction Text `json:"reaction"`
	// Actor is the display name of the participant who reacted.
	Actor Text `json:"actor"`
	// Timestamp is the Unix time of the reaction, in seconds. It is missing for
	// some reactions.
	Timestamp int64 `json:"timestamp,omitempty"`
}
