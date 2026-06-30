package dtd

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type RCONClient struct {
	addr     string
	password string
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer

	mu        sync.Mutex
	onMessage func(string)
	stopCh    chan struct{}
	wg        sync.WaitGroup
}

func NewRCONClient(addr, password string) *RCONClient {
	return &RCONClient{
		addr:     addr,
		password: password,
		stopCh:   make(chan struct{}),
	}
}

func (c *RCONClient) Connect(ctx context.Context) error {
	dialer := net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return fmt.Errorf("dial rcon %s: %w", c.addr, err)
	}
	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.writer = bufio.NewWriter(conn)

	if c.password != "" {
		if err := c.sendRaw(c.password); err != nil {
			_ = c.conn.Close()
			return fmt.Errorf("send password: %w", err)
		}
	}

	c.wg.Add(1)
	go c.readLoop()
	return nil
}

func (c *RCONClient) SendCommand(cmd string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.writer == nil {
		return fmt.Errorf("not connected")
	}
	_, err := c.writer.WriteString(cmd + "\r\n")
	if err != nil {
		return err
	}
	return c.writer.Flush()
}

func (c *RCONClient) sendRaw(s string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.writer == nil {
		return fmt.Errorf("not connected")
	}
	_, err := c.writer.WriteString(s + "\r\n")
	if err != nil {
		return err
	}
	return c.writer.Flush()
}

func (c *RCONClient) SetOnMessage(fn func(string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onMessage = fn
}

func (c *RCONClient) readLoop() {
	defer c.wg.Done()
	for {
		select {
		case <-c.stopCh:
			return
		default:
		}

		line, err := c.reader.ReadString('\n')
		if err != nil {
			c.mu.Lock()
			if c.onMessage != nil {
				c.onMessage(fmt.Sprintf("[error] read failed: %v", err))
			}
			c.mu.Unlock()
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		c.mu.Lock()
		if c.onMessage != nil {
			c.onMessage(line)
		}
		c.mu.Unlock()
	}
}

func (c *RCONClient) Close() error {
	close(c.stopCh)
	c.wg.Wait()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
