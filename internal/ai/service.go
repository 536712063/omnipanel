package ai

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	provider      Provider
	fileHandler   *FileContextHandler
	sessions      map[string]*Session
	sessionsMu    sync.RWMutex
	eventsEmitter EventsEmitter
}

type EventsEmitter interface {
	EmitEvent(eventName string, optionalData ...interface{})
}

type Session struct {
	ID       string
	Messages []ChatMessage
}

func NewService(provider Provider, emitter EventsEmitter) *Service {
	return &Service{
		provider:      provider,
		fileHandler:   NewFileContextHandler(),
		sessions:      make(map[string]*Session),
		eventsEmitter: emitter,
	}
}

func (s *Service) GetOrCreateSession(sessionID string) *Session {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	if session, ok := s.sessions[sessionID]; ok {
		return session
	}
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	session := &Session{ID: sessionID, Messages: []ChatMessage{}}
	s.sessions[sessionID] = session
	return session
}

func (s *Service) UploadFile(name string, data []byte) (ContentPart, error) {
	result, err := s.fileHandler.ProcessFile(name, data)
	if err != nil {
		return ContentPart{}, err
	}
	return result.Part, nil
}

func (s *Service) SendMessage(req ChatRequest) (ChatResponse, error) {
	ctx := context.Background()
	deltaCh := make(chan StreamEvent, 64)

	go func() {
		_ = s.provider.ChatStream(ctx, req, deltaCh)
		close(deltaCh)
	}()

	var content string
	for ev := range deltaCh {
		switch ev.Type {
		case "delta":
			content += ev.Delta
		case "error":
			return ChatResponse{}, fmt.Errorf(ev.Error)
		}
	}

	return ChatResponse{
		MessageID: uuid.New().String(),
		Content:   content,
		Done:      true,
	}, nil
}

func (s *Service) SendMessageStream(req ChatRequest) error {
	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
	}
	messageID := uuid.New().String()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	deltaCh := make(chan StreamEvent, 64)
	errCh := make(chan error, 1)

	go func() {
		errCh <- s.provider.ChatStream(ctx, req, deltaCh)
		close(deltaCh)
	}()

	go func() {
		for ev := range deltaCh {
			if s.eventsEmitter != nil {
				s.eventsEmitter.EmitEvent("ai:stream:event", ev)
			}
		}
		if err := <-errCh; err != nil && s.eventsEmitter != nil {
			s.eventsEmitter.EmitEvent("ai:stream:event", StreamEvent{
				Type:      "error",
				MessageID: messageID,
				Error:     err.Error(),
			})
		}
	}()

	return nil
}

func (s *Service) GetSessionHistory(sessionID string) []ChatMessage {
	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()
	if session, ok := s.sessions[sessionID]; ok {
		return session.Messages
	}
	return []ChatMessage{}
}

func (s *Service) SaveMessages(sessionID string, messages []ChatMessage) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	s.sessions[sessionID] = &Session{ID: sessionID, Messages: messages}
}
