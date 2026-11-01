package export

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

var discard = zap.NewNop()

const aliceFile = "testdata/messages/inbox/alice_123/message_1.json"

func TestRepairEncoding(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"ascii", "hello", "hello"},
		{"accent", "cafÃ©", "café"},
		{"emoji", "ð\u009f\u0098\u0082", "😂"},
		{"already valid UTF-8 above Latin-1", "café 😂", "café 😂"},
		{"Latin-1 characters that are not UTF-8 bytes", "café", "café"},
		{"empty", "", ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := repairEncoding(test.in); got != test.want {
				t.Errorf("repairEncoding(%q) = %q, want %q", test.in, got, test.want)
			}
		})
	}
}

func TestTextUnmarshalJSON(t *testing.T) {
	var got Text
	if err := json.Unmarshal([]byte(`"drÃ´le"`), &got); err != nil {
		t.Fatal(err)
	}
	if got != "drôle" {
		t.Errorf("got %q, want %q", got, "drôle")
	}
	if err := json.Unmarshal([]byte(`42`), &got); err == nil {
		t.Error("decoding a number into a Text: got no error")
	}
}

func TestReadFile(t *testing.T) {
	conversation, err := ReadFile(aliceFile)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := conversation.Title, Text("Alice Café"); got != want {
		t.Errorf("Title = %q, want %q", got, want)
	}
	if len(conversation.Participants) != 2 || conversation.Participants[0].Name != "Alice Café" {
		t.Errorf("Participants = %+v", conversation.Participants)
	}
	if conversation.ThreadPath != "inbox/alice_123" || !conversation.IsStillParticipant {
		t.Errorf("ThreadPath = %q, IsStillParticipant = %v", conversation.ThreadPath, conversation.IsStillParticipant)
	}
	if conversation.JoinableMode == nil || conversation.JoinableMode.Mode != 1 {
		t.Errorf("JoinableMode = %+v", conversation.JoinableMode)
	}
	if conversation.IsPending != nil {
		t.Errorf("IsPending = %v, want nil", *conversation.IsPending)
	}

	messages := conversation.Messages
	if len(messages) != 5 {
		t.Fatalf("got %d messages, want 5", len(messages))
	}
	wantKinds := []Kind{KindText, KindShare, KindMedia, KindAudio, KindCall}
	for i, want := range wantKinds {
		if got := messages[i].Kind(); got != want {
			t.Errorf("message %d: Kind() = %v, want %v", i, got, want)
		}
	}

	text := messages[0]
	if text.Content != "😂 trop drôle" {
		t.Errorf("Content = %q", text.Content)
	}
	if len(text.Reactions) != 1 || text.Reactions[0] != (Reaction{Reaction: "❤", Actor: "Alice Café", Timestamp: 1700000400}) {
		t.Errorf("Reactions = %+v", text.Reactions)
	}
	if want := time.UnixMilli(1700000300000); !text.Time().Equal(want) {
		t.Errorf("Time() = %v, want %v", text.Time(), want)
	}

	share := messages[1].Share
	if share == nil || share.Link != "https://www.instagram.com/reel/abc/" || share.OriginalContentOwner != "someone" {
		t.Errorf("Share = %+v", share)
	}
	media := messages[2]
	if len(media.Photos) != 1 || media.Photos[0].CreationTimestamp != 1700000100 || len(media.Videos) != 1 {
		t.Errorf("Photos = %+v, Videos = %+v", media.Photos, media.Videos)
	}
	if call := messages[4].CallDuration; call == nil || *call != 42 {
		t.Errorf("CallDuration = %v, want 42", call)
	}
}

func TestReadFileErrors(t *testing.T) {
	if _, err := ReadFile("testdata/messages/secret_conversations.json"); !errors.Is(err, ErrNotConversation) {
		t.Errorf("secret conversations: got error %v, want ErrNotConversation", err)
	}
	if _, err := ReadFile("testdata/messages/inbox/broken_456/message_1.json"); err == nil {
		t.Error("invalid JSON: got no error")
	}
	if _, err := ReadFile("testdata/missing.json"); err == nil {
		t.Error("missing file: got no error")
	}
}

func TestLoad(t *testing.T) {
	t.Run("folder of conversations", func(t *testing.T) {
		// The secret conversations, the invalid file and the photo are skipped.
		conversations, err := Load([]string{"testdata/messages"}, discard)
		if err != nil {
			t.Fatal(err)
		}
		if len(conversations) != 1 || conversations[0].ThreadPath != "inbox/alice_123" {
			t.Errorf("got %d conversations, want the one of alice_123", len(conversations))
		}
	})

	t.Run("files and folders", func(t *testing.T) {
		// The file is in the folder too, it is read once.
		conversations, err := Load([]string{aliceFile, filepath.Dir(aliceFile), "testdata/messages"}, discard)
		if err != nil {
			t.Fatal(err)
		}
		if len(conversations) != 1 {
			t.Errorf("got %d conversations, want 1", len(conversations))
		}
	})

	t.Run("unreadable subfolder", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("root can list any folder")
		}
		dir := t.TempDir()
		data, err := os.ReadFile(aliceFile)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "message_1.json"), data, 0o600); err != nil {
			t.Fatal(err)
		}
		locked := filepath.Join(dir, "locked")
		if err := os.Mkdir(locked, 0o000); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(locked, 0o700) })

		conversations, err := Load([]string{dir}, discard)
		if err != nil {
			t.Fatal(err)
		}
		if len(conversations) != 1 {
			t.Errorf("got %d conversations, want 1", len(conversations))
		}
	})

	t.Run("a file that is not a conversation", func(t *testing.T) {
		if _, err := Load([]string{"testdata/messages/secret_conversations.json"}, discard); !errors.Is(err, ErrNotConversation) {
			t.Errorf("got error %v, want ErrNotConversation", err)
		}
	})

	t.Run("missing source", func(t *testing.T) {
		if _, err := Load([]string{"testdata/missing"}, discard); err == nil {
			t.Error("got no error")
		}
	})
}

func TestTextString(t *testing.T) {
	if got := Text("héllo").String(); got != "héllo" {
		t.Fatalf("String() = %q", got)
	}
}
