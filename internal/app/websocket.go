package app

import (
	"context"
	"sync"
	"time"
)

// WebSocketHub 管理所有 WebSocket 连接并广播消息。
// 用于实时推送:
//   - 七日杀控制台日志
//   - Docker 容器日志
//   - 系统监控数据
//   - 远程终端输出
type WebSocketHub struct {
	clients    map[string]map[*Client]bool // sessionID -> clients
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// Client 表示一个 WebSocket 连接
type Client struct {
	ID        string
	SessionID string
	Send      chan []byte
}

// Message 广播消息
type Message struct {
	SessionID string
	Data      []byte
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run 启动 Hub 的主事件循环。
func (h *WebSocketHub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.SessionID] == nil {
				h.clients[client.SessionID] = make(map[*Client]bool)
			}
			h.clients[client.SessionID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.SessionID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.SessionID)
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients, ok := h.clients[msg.SessionID]
			h.mu.RUnlock()
			if ok {
				for client := range clients {
					select {
					case client.Send <- msg.Data:
					default:
						h.mu.Lock()
						delete(h.clients[msg.SessionID], client)
						close(client.Send)
						h.mu.Unlock()
					}
				}
			}
		}
	}
}

// Broadcast 广播消息到指定 session 的所有客户端。
func (h *WebSocketHub) Broadcast(sessionID string, data interface{}) {
	var jsonData []byte
	switch v := data.(type) {
	case []byte:
		jsonData = v
	case string:
		jsonData = []byte(v)
	case map[string]string:
		// 简单 JSON 序列化
		jsonData = []byte(`{"status":"` + v["status"] + `","message":"` + v["message"] + `"}`)
	default:
		return
	}

	h.broadcast <- &Message{
		SessionID: sessionID,
		Data:      jsonData,
	}
}

// Shutdown 关闭所有 WebSocket 连接。
func (h *WebSocketHub) Shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, clients := range h.clients {
		for client := range clients {
			delete(clients, client)
			close(client.Send)
		}
	}
	h.clients = make(map[string]map[*Client]bool)
}

// RegisterClient 注册一个 WebSocket 客户端。
func (h *WebSocketHub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient 注销一个 WebSocket 客户端。
func (h *WebSocketHub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// Ticker 用于定期发送心跳或其他周期性数据
func (h *WebSocketHub) StartTicker(ctx context.Context, sessionID string, interval time.Duration, getData func() interface{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.Broadcast(sessionID, getData())
		}
	}
}
