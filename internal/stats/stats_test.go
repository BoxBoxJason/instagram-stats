package stats

import (
	"reflect"
	"testing"
	"time"

	"github.com/boxboxjason/instagram-stats/internal/export"
)

// at returns the timestamp in milliseconds of a time in UTC.
func at(year int, month time.Month, day, hour int) int64 {
	return time.Date(year, month, day, hour, 0, 0, 0, time.UTC).UnixMilli()
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func message(sender string, timestamp int64, content string) export.Message {
	return export.Message{SenderName: export.Text(sender), TimestampMs: timestamp, Content: export.Text(content)}
}

func TestAdd(t *testing.T) {
	duration := int64(60)
	share := message("Bob", at(2024, 1, 1, 10), "Bob sent an attachment.")
	share.Share = &export.Share{Link: "https://www.instagram.com/reel/abc/"}
	media := message("Bob", at(2024, 1, 1, 11), "")
	media.Photos = []export.Media{{URI: "1.jpg"}}
	media.Videos = []export.Media{{URI: "1.mp4"}}
	call := message("Alice", at(2024, 1, 1, 12), "Audio call ended")
	call.CallDuration = &duration
	reacted := message("Alice", at(2024, 1, 1, 13), "LOL, it's Alice's cat!")
	reacted.Reactions = []export.Reaction{{Reaction: "❤", Actor: "Bob"}, {Reaction: "😂", Actor: "Ghost"}}

	s := New(time.UTC, []string{"LOL", "cat", "dog", ""})
	s.Add(&export.Conversation{
		Participants: []export.Participant{{Name: "Alice"}, {Name: "Bob"}, {Name: "Carol"}},
		Messages: []export.Message{
			reacted,
			call,
			media,
			share,
			message("Alice", at(2024, 1, 1, 14), "Liked a message"),
			message("Alice", at(2024, 1, 1, 15), "Alice sent an attachment."),
			message("Dave", at(2024, 1, 1, 16), "hello lol"),
		},
	})

	type counts struct {
		Messages, Reactions, Reels, PhotosAndVideos, WordsFound, Vocabulary int
	}
	want := map[string]counts{
		// The call and the placeholder are messages, but their words are not counted.
		"Alice": {Messages: 3, WordsFound: 2, Vocabulary: 4},
		"Bob":   {Reactions: 1, Reels: 1, PhotosAndVideos: 1},
		"Carol": {},
		// Dave and Ghost are not participants anymore.
		"Dave":  {Messages: 1, WordsFound: 1, Vocabulary: 2},
		"Ghost": {Reactions: 1},
	}
	var names []string
	for _, p := range s.Participants() {
		names = append(names, p.Name)
		got := counts{p.Messages, p.Reactions, p.Reels, p.PhotosAndVideos, p.WordsFound, p.Vocabulary()}
		if got != want[p.Name] {
			t.Errorf("%s: got %+v, want %+v", p.Name, got, want[p.Name])
		}
	}
	if wantNames := []string{"Alice", "Bob", "Carol", "Ghost", "Dave"}; !reflect.DeepEqual(names, wantNames) {
		t.Errorf("participants = %v, want %v", names, wantNames)
	}
	// lol, it's, alice's, cat, hello: "lol" is shared by Alice and Dave.
	if got := s.VocabularySize(); got != 5 {
		t.Errorf("VocabularySize() = %d, want 5", got)
	}
	// Every message but the like is dated.
	if days := s.Days(); len(days) != 1 || days[0].Messages != 6 {
		t.Errorf("Days() = %+v, want 6 messages on a single day", days)
	}
}

func TestWords(t *testing.T) {
	tests := map[string][]string{
		"hello world":             {"hello", "world"},
		"  Hello,   world!! ":     {"Hello", "world"},
		"j'ai vu l’été 2024":      {"j'ai", "vu", "l’été", "2024"},
		"'quoted' -- ...":         {"quoted"},
		"😂😂 ok":                   {"ok"},
		"":                        nil,
		"https://example.com/a?b": {"https", "example", "com", "a", "b"},
	}
	for text, want := range tests {
		got := Words(text)
		if len(got) == 0 && len(want) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Words(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestTimeStatistics(t *testing.T) {
	s := New(time.UTC, nil)
	s.Add(&export.Conversation{Messages: []export.Message{
		// Messages come from the most recent, as in the export.
		message("Alice", at(2024, 3, 2, 9), "c"),
		message("Alice", at(2024, 2, 28, 23), "b"),
		message("Alice", at(2024, 2, 28, 22), "b"),
		message("Alice", at(2024, 1, 29, 8), "a"),
	}})

	days := s.Days()
	// From January 29 to March 2, both included, 2024 being a leap year.
	if len(days) != 34 {
		t.Fatalf("got %d days, want 34", len(days))
	}
	if days[0] != (Count{Start: date(2024, 1, 29), Messages: 1}) || days[33] != (Count{Start: date(2024, 3, 2), Messages: 1}) {
		t.Errorf("first and last days = %+v, %+v", days[0], days[33])
	}
	if days[1] != (Count{Start: date(2024, 1, 30)}) {
		t.Errorf("day without messages = %+v", days[1])
	}

	wantMonths := []Count{
		{Start: date(2024, 1, 1), Messages: 1},
		{Start: date(2024, 2, 1), Messages: 2},
		{Start: date(2024, 3, 1), Messages: 1},
	}
	if months := s.Months(); !reflect.DeepEqual(months, wantMonths) {
		t.Errorf("Months() = %+v, want %+v", months, wantMonths)
	}

	// The 34 days hold 5 Wednesdays: February 28 2024 is one.
	week := s.AverageWeek()
	if got := week[2][22]; got != 1.0/5 {
		t.Errorf("Wednesday 22:00 = %v, want %v", got, 1.0/5)
	}
	// January 29 2024 is a Monday, among the 5 Mondays.
	if got := week[0][8]; got != 1.0/5 {
		t.Errorf("Monday 08:00 = %v, want %v", got, 1.0/5)
	}
	if got := week[0][9]; got != 0 {
		t.Errorf("Monday 09:00 = %v, want 0", got)
	}
}

func TestTimeZone(t *testing.T) {
	tokyo := time.FixedZone("UTC+9", 9*3600)
	s := New(tokyo, nil)
	// 20:00 UTC on a Sunday is 05:00 on the Monday in UTC+9.
	s.Add(&export.Conversation{Messages: []export.Message{message("Alice", at(2024, 1, 7, 20), "hi")}})
	if days := s.Days(); len(days) != 1 || !days[0].Start.Equal(date(2024, 1, 8)) {
		t.Errorf("Days() = %+v, want January 8", days)
	}
	if got := s.AverageWeek()[0][5]; got != 1 {
		t.Errorf("Monday 05:00 = %v, want 1", got)
	}
}

func TestNoMessages(t *testing.T) {
	s := New(time.UTC, nil)
	s.Add(&export.Conversation{Participants: []export.Participant{{Name: "Alice"}}})
	if days, months := s.Days(), s.Months(); days != nil || months != nil {
		t.Errorf("Days() = %v, Months() = %v, want none", days, months)
	}
	if week := s.AverageWeek(); week != ([7][24]float64{}) {
		t.Errorf("AverageWeek() = %v, want zeros", week)
	}
}

func TestRanking(t *testing.T) {
	s := New(time.UTC, nil)
	for _, name := range []string{"Alice", "Bob", "Carol", "Dave"} {
		s.participant(name)
	}
	s.byName["Alice"].Messages = 2
	s.byName["Bob"].Messages = 5
	s.byName["Carol"].Messages = 0
	s.byName["Dave"].Messages = 2

	want := []Rank{
		{Count: 5, Names: []string{"Bob"}},
		{Count: 2, Names: []string{"Alice", "Dave"}},
		{Count: 0, Names: []string{"Carol"}},
	}
	if got := s.Ranking(func(p *Participant) int { return p.Messages }); !reflect.DeepEqual(got, want) {
		t.Errorf("Ranking() = %+v, want %+v", got, want)
	}
}
