// Package stats computes the statistics of Instagram conversations: what each
// participant sent, and when the messages were sent.
package stats

import (
	"cmp"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/boxboxjason/instagram-stats/internal/export"
)

// daysInWeek and hoursInDay are the dimensions of the grid of a week.
const (
	daysInWeek = 7
	hoursInDay = 24
)

// likedMessage is the content of the messages Instagram logs for a like, which are
// not messages sent by the participant.
const likedMessage = "Liked a message"

// attachmentPlaceholder ends the content Instagram writes for a message whose
// attachment it did not export, such as "Alice sent an attachment.".
const attachmentPlaceholder = " sent an attachment."

// Participant holds what a participant sent.
type Participant struct {
	vocabulary      map[string]struct{}
	Name            string
	Messages        int
	Reactions       int
	Reels           int
	PhotosAndVideos int
	WordsFound      int
}

// Vocabulary returns the number of different words the participant used.
func (p *Participant) Vocabulary() int {
	return len(p.vocabulary)
}

// Stats accumulates the statistics of conversations. The zero value is not usable,
// create one with [New].
type Stats struct {
	location     *time.Location
	byName       map[string]*Participant
	days         map[time.Time]int
	searched     []string
	participants []*Participant
	week         [daysInWeek][hoursInDay]int
}

// New returns empty statistics. Messages are dated in location, and the words
// of searched are counted in them regardless of case.
func New(location *time.Location, searched []string) *Stats {
	lowered := make([]string, 0, len(searched))
	for _, word := range searched {
		if word != "" {
			lowered = append(lowered, strings.ToLower(word))
		}
	}

	return &Stats{
		location: location,
		searched: lowered,
		byName:   map[string]*Participant{},
		days:     map[time.Time]int{},
	}
}

// Add adds the participants and messages of a conversation to the statistics.
func (s *Stats) Add(conversation *export.Conversation) {
	for _, participant := range conversation.Participants {
		s.participant(participant.Name.String())
	}

	for _, message := range conversation.Messages {
		s.addMessage(message)
	}
}

// Words splits text into its words: the runs of letters, digits and apostrophes,
// without the apostrophes around them.
func Words(text string) []string {
	fields := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && !isApostrophe(r)
	})

	words := fields[:0]
	for _, field := range fields {
		if word := strings.TrimFunc(field, isApostrophe); word != "" {
			words = append(words, word)
		}
	}

	return words
}

func isApostrophe(r rune) bool {
	return r == '\'' || r == '’'
}

// Participants returns the participants, in the order they were met.
func (s *Stats) Participants() []*Participant {
	return s.participants
}

// VocabularySize returns the number of different words used by all participants.
func (s *Stats) VocabularySize() int {
	all := map[string]struct{}{}

	for _, p := range s.participants {
		for word := range p.vocabulary {
			all[word] = struct{}{}
		}
	}

	return len(all)
}

// Count is a number of messages sent during a period.
type Count struct {
	// Start is the first day of the period, at midnight UTC.
	Start    time.Time
	Messages int
}

// Days returns the number of messages of every day, from the day of the first message
// to the day of the last one, both included. It is empty when there are no messages.
func (s *Stats) Days() []Count {
	first, last, ok := s.span()
	if !ok {
		return nil
	}

	var days []Count
	for day := first; !day.After(last); day = day.AddDate(0, 0, 1) {
		days = append(days, Count{Start: day, Messages: s.days[day]})
	}

	return days
}

// Months returns the number of messages of every month, from the month of the first
// message to the month of the last one, both included.
func (s *Stats) Months() []Count {
	var months []Count

	for _, day := range s.Days() {
		month := time.Date(day.Start.Year(), day.Start.Month(), 1, 0, 0, 0, 0, time.UTC)
		if len(months) == 0 || !months[len(months)-1].Start.Equal(month) {
			months = append(months, Count{Start: month})
		}

		months[len(months)-1].Messages += day.Messages
	}

	return months
}

// AverageWeek returns the average number of messages sent on each day of the week,
// from Monday, and each hour of the day, over the weeks from the first message to
// the last one.
func (s *Stats) AverageWeek() [daysInWeek][hoursInDay]float64 {
	var average [daysInWeek][hoursInDay]float64

	first, last, ok := s.span()
	if !ok {
		return average
	}

	var occurrences [daysInWeek]int
	for day := first; !day.After(last); day = day.AddDate(0, 0, 1) {
		occurrences[weekdayIndex(day.Weekday())]++
	}

	for weekday, hours := range s.week {
		for hour, messages := range hours {
			if occurrences[weekday] > 0 {
				average[weekday][hour] = float64(messages) / float64(occurrences[weekday])
			}
		}
	}

	return average
}

// Rank is a position of a ranking: the participants sharing a count.
type Rank struct {
	Names []string
	Count int
}

// Ranking ranks the participants by decreasing count.
func (s *Stats) Ranking(count func(*Participant) int) []Rank {
	var ranking []Rank

	byCount := map[int]int{}

	for _, participant := range s.participants {
		total := count(participant)
		if i, ok := byCount[total]; ok {
			ranking[i].Names = append(ranking[i].Names, participant.Name)

			continue
		}

		byCount[total] = len(ranking)
		ranking = append(ranking, Rank{Count: total, Names: []string{participant.Name}})
	}

	slices.SortStableFunc(ranking, func(a, b Rank) int { return cmp.Compare(b.Count, a.Count) })

	return ranking
}

// participant returns the participant named name, added if it is not known yet.
func (s *Stats) participant(name string) *Participant {
	if p, ok := s.byName[name]; ok {
		return p
	}

	p := &Participant{Name: name, vocabulary: map[string]struct{}{}}
	s.byName[name] = p
	s.participants = append(s.participants, p)

	return p
}

func (s *Stats) addMessage(message export.Message) {
	content := message.Content.String()
	if content == likedMessage {
		return
	}

	sender := s.participant(message.SenderName.String())
	switch kind := message.Kind(); kind {
	case export.KindShare:
		sender.Reels++
	case export.KindMedia:
		sender.PhotosAndVideos++
	case export.KindText, export.KindAudio, export.KindCall:
		sender.Messages++
		if kind == export.KindText && content != "" && !strings.HasSuffix(content, attachmentPlaceholder) {
			s.addText(sender, content)
		}
	}

	for _, reaction := range message.Reactions {
		s.participant(reaction.Actor.String()).Reactions++
	}

	sent := message.Time().In(s.location)
	s.days[dateOf(sent)]++
	s.week[weekdayIndex(sent.Weekday())][sent.Hour()]++
}

func (s *Stats) addText(sender *Participant, content string) {
	lowered := strings.ToLower(content)
	for _, word := range s.searched {
		if strings.Contains(lowered, word) {
			sender.WordsFound++
		}
	}

	for _, word := range Words(lowered) {
		sender.vocabulary[word] = struct{}{}
	}
}

// span returns the days of the first and of the last message.
func (s *Stats) span() (time.Time, time.Time, bool) {
	var (
		first, last time.Time
		found       bool
	)

	for day := range s.days {
		if !found || day.Before(first) {
			first = day
		}

		if !found || day.After(last) {
			last = day
		}

		found = true
	}

	return first, last, found
}

// dateOf returns the calendar day of t, at midnight UTC.
func dateOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// weekdayIndex returns the index of weekday in a week starting on Monday.
func weekdayIndex(weekday time.Weekday) int {
	return (int(weekday) + daysInWeek - 1) % daysInWeek
}
