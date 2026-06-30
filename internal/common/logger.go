package common

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
	LevelFatal LogLevel = "fatal"
)

type AppError struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Error     string                 `json:"error,omitempty"`
	Stack     string                 `json:"stack,omitempty"`
	Component string                 `json:"component"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

type Logger struct {
	logDir      string
	mu          sync.Mutex
	currentFile *os.File
	writer      io.Writer
	minLevel    LogLevel
}

func NewLogger(logDir string) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}
	logger := &Logger{
		logDir:   logDir,
		minLevel: LevelInfo,
	}
	if err := logger.rotate(); err != nil {
		return nil, err
	}
	logger.writer = io.MultiWriter(os.Stdout, logger.currentFile)
	log.SetOutput(logger.writer)
	return logger, nil
}

func (l *Logger) rotate() error {
	date := time.Now().Format("2006-01-02")
	path := filepath.Join(l.logDir, fmt.Sprintf("omnipanel-%s.log", date))
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if l.currentFile != nil {
		_ = l.currentFile.Close()
	}
	l.currentFile = f
	if l.writer != nil {
		l.writer = io.MultiWriter(os.Stdout, f)
		log.SetOutput(l.writer)
	}
	return nil
}

func (l *Logger) SetMinLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
}

func levelRank(level LogLevel) int {
	switch level {
	case LevelDebug:
		return 0
	case LevelInfo:
		return 1
	case LevelWarn:
		return 2
	case LevelError:
		return 3
	case LevelFatal:
		return 4
	}
	return 1
}

func (l *Logger) Log(level LogLevel, component, message string, err error, ctx map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if levelRank(level) < levelRank(l.minLevel) {
		return
	}
	_ = l.rotate()

	entry := AppError{
		ID:        generateID(),
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Component: component,
		Context:   ctx,
	}
	if err != nil {
		entry.Error = err.Error()
		entry.Stack = fmt.Sprintf("%+v", err)
	}

	data, _ := json.Marshal(entry)
	log.Println(string(data))
}

func (l *Logger) Debug(component, message string, ctx map[string]interface{}) {
	l.Log(LevelDebug, component, message, nil, ctx)
}
func (l *Logger) Info(component, message string, ctx map[string]interface{}) {
	l.Log(LevelInfo, component, message, nil, ctx)
}
func (l *Logger) Warn(component, message string, err error, ctx map[string]interface{}) {
	l.Log(LevelWarn, component, message, err, ctx)
}
func (l *Logger) Error(component, message string, err error, ctx map[string]interface{}) {
	l.Log(LevelError, component, message, err, ctx)
}
func (l *Logger) Fatal(component, message string, err error, ctx map[string]interface{}) {
	l.Log(LevelFatal, component, message, err, ctx)
}

func (l *Logger) ReadLogs(days int, limit int) ([]AppError, error) {
	if days <= 0 {
		days = 7
	}
	if limit <= 0 {
		limit = 100
	}

	var results []AppError
	now := time.Now()
	for d := 0; d < days && len(results) < limit; d++ {
		date := now.AddDate(0, 0, -d).Format("2006-01-02")
		path := filepath.Join(l.logDir, fmt.Sprintf("omnipanel-%s.log", date))
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		lines := splitLines(string(data))
		for _, line := range lines {
			if line == "" {
				continue
			}
			var entry AppError
			if json.Unmarshal([]byte(line), &entry) == nil {
				results = append(results, entry)
				if len(results) >= limit {
					break
				}
			}
		}
	}
	return results, nil
}

func generateID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func GlobalRecovery(logger *Logger, next func()) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("global", "panic recovered", fmt.Errorf("%v", r), map[string]interface{}{
				"recover": r,
			})
		}
	}()
	next()
}
