package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sync"
)

type Agent interface {
	Name() string
	Version() string
	Description() string
	Execute(ctx map[string]interface{}) (map[string]interface{}, error)
}

type Manifest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Entry       string            `json:"entry"`
	Type        string            `json:"type"`
	Permissions []string          `json:"permissions"`
	Config      map[string]string `json:"config"`
}

type Plugin struct {
	Manifest Manifest
	Agent    Agent
	Path     string
}

type Marketplace struct {
	mu       sync.RWMutex
	plugins  map[string]*Plugin
	registry []Manifest
	loadDir  string
}

func NewMarketplace(loadDir string) *Marketplace {
	return &Marketplace{
		plugins: make(map[string]*Plugin),
		loadDir: loadDir,
	}
}

func (m *Marketplace) RegisterBuiltIn(agent Agent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := fmt.Sprintf("builtin-%s", agent.Name())
	if _, exists := m.plugins[id]; exists {
		return fmt.Errorf("plugin already registered: %s", id)
	}

	p := &Plugin{
		Manifest: Manifest{
			ID:          id,
			Name:        agent.Name(),
			Version:     agent.Version(),
			Description: agent.Description(),
			Type:        "agent",
		},
		Agent: agent,
	}

	m.plugins[id] = p
	m.refreshRegistry()
	return nil
}

func (m *Marketplace) LoadFromDir() error {
	if m.loadDir == "" {
		return nil
	}

	if err := os.MkdirAll(m.loadDir, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(m.loadDir)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, entry := range entries {
		if entry.IsDir() {
			manifestPath := filepath.Join(m.loadDir, entry.Name(), "manifest.json")
			soPath := filepath.Join(m.loadDir, entry.Name(), fmt.Sprintf("%s.so", entry.Name()))
			_ = m.loadPlugin(manifestPath, soPath)
		}
	}
	m.refreshRegistry()
	return nil
}

func (m *Marketplace) loadPlugin(manifestPath string, soPath string) error {
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	var manifest Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return err
	}

	p, err := plugin.Open(soPath)
	if err != nil {
		return err
	}

	sym, err := p.Lookup("Agent")
	if err != nil {
		return err
	}

	agent, ok := sym.(Agent)
	if !ok {
		return fmt.Errorf("plugin does not implement Agent interface")
	}

	m.plugins[manifest.ID] = &Plugin{
		Manifest: manifest,
		Agent:    agent,
		Path:     soPath,
	}
	return nil
}

func (m *Marketplace) refreshRegistry() {
	m.registry = make([]Manifest, 0, len(m.plugins))
	for _, p := range m.plugins {
		m.registry = append(m.registry, p.Manifest)
	}
}

func (m *Marketplace) ListPlugins() []Manifest {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]Manifest, len(m.registry))
	copy(result, m.registry)
	return result
}

func (m *Marketplace) GetPlugin(id string) (*Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.plugins[id]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", id)
	}
	return p, nil
}

func (m *Marketplace) ExecuteAgent(id string, ctx map[string]interface{}) (map[string]interface{}, error) {
	p, err := m.GetPlugin(id)
	if err != nil {
		return nil, err
	}
	if p.Agent == nil {
		return nil, fmt.Errorf("agent not available")
	}
	return p.Agent.Execute(ctx)
}

func (m *Marketplace) DisablePlugin(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.plugins, id)
	m.refreshRegistry()
}

func (m *Marketplace) EnablePlugin(id string) {
	m.refreshRegistry()
}
