package dtd

import (
	"regexp"
	"strings"
	"time"
)

type LogEvent string

const (
	EventPlayerConnected    LogEvent = "player_connected"
	EventPlayerDisconnected LogEvent = "player_disconnected"
	EventPlayerSpawned      LogEvent = "player_spawned"
	EventChat               LogEvent = "chat"
	EventServerStarted      LogEvent = "server_started"
	EventBloodMoonWarning   LogEvent = "blood_moon_warning"
	EventUnknown            LogEvent = "unknown"
)

type ParsedLogEvent struct {
	Timestamp  time.Time
	Type       LogEvent
	PlayerID   string
	PlayerName string
	Message    string
	Raw        string
}

type LogParser struct {
	patterns []logPattern
}

type logPattern struct {
	re      *regexp.Regexp
	extract func(matches []string) ParsedLogEvent
	event   LogEvent
}

func NewLogParser() *LogParser {
	return &LogParser{
		patterns: []logPattern{
			{
				re: regexp.MustCompile(`Player '(.+?)' joined the game`),
				extract: func(m []string) ParsedLogEvent {
					return ParsedLogEvent{PlayerName: m[1], Message: m[0]}
				},
				event: EventPlayerConnected,
			},
			{
				re: regexp.MustCompile(`Player '(.+?)' left the game`),
				extract: func(m []string) ParsedLogEvent {
					return ParsedLogEvent{PlayerName: m[1], Message: m[0]}
				},
				event: EventPlayerDisconnected,
			},
			{
				re: regexp.MustCompile(`PlayerSpawnedInWorld:.*?PlayerName='(.+?)'`),
				extract: func(m []string) ParsedLogEvent {
					return ParsedLogEvent{PlayerName: m[1], Message: m[0]}
				},
				event: EventPlayerSpawned,
			},
			{
				re: regexp.MustCompile(`Chat: '(.+?)':\s*(.+)`),
				extract: func(m []string) ParsedLogEvent {
					return ParsedLogEvent{PlayerName: m[1], Message: m[2]}
				},
				event: EventChat,
			},
			{
				re: regexp.MustCompile(`(?i)blood\s*moon.*?(rising|soon|tonight|started)`),
				extract: func(m []string) ParsedLogEvent {
					return ParsedLogEvent{Message: m[0]}
				},
				event: EventBloodMoonWarning,
			},
		},
	}
}

func (p *LogParser) ParseLine(raw string) ParsedLogEvent {
	parsed := ParsedLogEvent{Type: EventUnknown, Raw: raw}
	parsed.Timestamp = extractTimestamp(raw)

	for _, pat := range p.patterns {
		matches := pat.re.FindStringSubmatch(raw)
		if matches == nil {
			continue
		}
		event := pat.extract(matches)
		event.Type = pat.event
		event.Timestamp = parsed.Timestamp
		event.Raw = raw
		return event
	}

	return parsed
}

func extractTimestamp(raw string) time.Time {
	re := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})`)
	matches := re.FindStringSubmatch(raw)
	if len(matches) > 1 {
		if t, err := time.Parse("2006-01-02 15:04:05", matches[1]); err == nil {
			return t
		}
	}
	return time.Now()
}

type PlayerListParser struct{}

func NewPlayerListParser() *PlayerListParser {
	return &PlayerListParser{}
}

func (p *PlayerListParser) Parse(output string) []OnlinePlayer {
	var players []OnlinePlayer
	lines := strings.Split(output, "\n")
	re := regexp.MustCompile(`(?i)id\s*=\s*(\d+).*?name\s*=\s*([^,]+)`)

	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		players = append(players, OnlinePlayer{
			ID:         matches[1],
			Name:       strings.TrimSpace(matches[2]),
			Status:     "online",
			JoinedAt:   time.Now().Format(time.RFC3339),
			PlatformID: "",
		})
	}
	return players
}
