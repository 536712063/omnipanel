package dtd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	mu           sync.RWMutex
	servers      map[string]*ServerInstance
	emitter      EventsEmitter
	multiMachine MultiMachineBridge
	scheduler    *Scheduler
	logParser    *LogParser
}

type ServerInstance struct {
	Config     ServerConfig
	RCON       *RCONClient
	Players    map[string]OnlinePlayer
	Events     []PlayerEvent
	ConsoleBuf []ConsoleMessage
	MapViewer  *MapViewer
	running    bool
}

func NewService(emitter EventsEmitter, bridge MultiMachineBridge) *Service {
	svc := &Service{
		servers:      make(map[string]*ServerInstance),
		emitter:      emitter,
		multiMachine: bridge,
		logParser:    NewLogParser(),
	}
	svc.scheduler = NewScheduler(svc)
	svc.scheduler.Start()
	return svc
}

func (s *Service) Stop() {
	if s.scheduler != nil {
		s.scheduler.Stop()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, inst := range s.servers {
		if inst.RCON != nil {
			_ = inst.RCON.Close()
		}
	}
}

func (s *Service) AddServer(cfg ServerConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cfg.ID == "" {
		cfg.ID = uuid.New().String()
	}
	if cfg.BloodMoonDay == 0 {
		cfg.BloodMoonDay = 7
	}
	inst := &ServerInstance{
		Config:     cfg,
		Players:    make(map[string]OnlinePlayer),
		Events:     []PlayerEvent{},
		ConsoleBuf: make([]ConsoleMessage, 0, 500),
		MapViewer:  NewMapViewer(""),
	}
	s.servers[cfg.ID] = inst
	return nil
}

func (s *Service) GetServer(id string) (*ServerInstance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	inst, ok := s.servers[id]
	if !ok {
		return nil, fmt.Errorf("server not found")
	}
	copyInst := *inst
	return &copyInst, nil
}

func (s *Service) ListServers() []ServerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]ServerConfig, 0, len(s.servers))
	for _, inst := range s.servers {
		result = append(result, inst.Config)
	}
	return result
}

func (s *Service) Connect(serverID string) error {
	inst, err := s.GetServer(serverID)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", inst.Config.Host, inst.Config.RCONPort)
	client := NewRCONClient(addr, inst.Config.RCONPass)
	if err := client.Connect(context.Background()); err != nil {
		return err
	}

	client.SetOnMessage(func(line string) {
		s.handleRCONMessage(serverID, line)
	})

	s.mu.Lock()
	server := s.servers[serverID]
	server.RCON = client
	server.running = true
	s.mu.Unlock()

	return s.RefreshPlayers(serverID)
}

func (s *Service) Disconnect(serverID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	inst, ok := s.servers[serverID]
	if !ok {
		return fmt.Errorf("server not found")
	}
	if inst.RCON != nil {
		if err := inst.RCON.Close(); err != nil {
			return err
		}
	}
	inst.RCON = nil
	inst.running = false
	return nil
}

func (s *Service) SendCommand(serverID string, cmd string) error {
	s.mu.RLock()
	inst := s.servers[serverID]
	s.mu.RUnlock()
	if inst == nil || inst.RCON == nil {
		return fmt.Errorf("server not connected")
	}
	return inst.RCON.SendCommand(cmd)
}

func (s *Service) RefreshPlayers(serverID string) error {
	if err := s.SendCommand(serverID, "lp"); err != nil {
		return err
	}
	return nil
}

func (s *Service) GetPlayers(serverID string) []OnlinePlayer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	inst := s.servers[serverID]
	if inst == nil {
		return []OnlinePlayer{}
	}
	result := make([]OnlinePlayer, 0, len(inst.Players))
	for _, p := range inst.Players {
		result = append(result, p)
	}
	return result
}

func (s *Service) GetPlayerEvents(serverID string, limit int) []PlayerEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	inst := s.servers[serverID]
	if inst == nil {
		return []PlayerEvent{}
	}
	if limit <= 0 || limit > len(inst.Events) {
		limit = len(inst.Events)
	}
	return inst.Events[len(inst.Events)-limit:]
}

func (s *Service) GetConsoleMessages(serverID string, limit int) []ConsoleMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	inst := s.servers[serverID]
	if inst == nil {
		return []ConsoleMessage{}
	}
	if limit <= 0 || limit > len(inst.ConsoleBuf) {
		limit = len(inst.ConsoleBuf)
	}
	return inst.ConsoleBuf[len(inst.ConsoleBuf)-limit:]
}

func (s *Service) ParseLogFile(serverID string, logPath string) error {
	file, err := os.Open(logPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		s.handleRCONMessage(serverID, line)
	}
	return scanner.Err()
}

func (s *Service) GetMapData(serverID string) (*MapData, error) {
	s.mu.RLock()
	inst := s.servers[serverID]
	s.mu.RUnlock()
	if inst == nil {
		return nil, fmt.Errorf("server not found")
	}
	if inst.Config.SaveDir == "" {
		return nil, fmt.Errorf("save_dir not configured")
	}
	return inst.MapViewer.LoadFromSaveDir(inst.Config.SaveDir)
}

func (s *Service) GetBloodMoonInfo(serverID string) (*BloodMoonInfo, error) {
	currentDay := 1
	next := CalculateBloodMoon(currentDay, 7)
	return &BloodMoonInfo{
		CurrentDay:    currentDay,
		NextBloodMoon: next,
		RemainingDays: next - currentDay,
		RemainingHours: (next - currentDay) * 24,
	}, nil
}

func (s *Service) AddScheduledTask(task ScheduledTask) (*ScheduledTask, error) {
	return s.scheduler.AddTask(task)
}

func (s *Service) RemoveScheduledTask(taskID string) {
	s.scheduler.RemoveTask(taskID)
}

func (s *Service) GetScheduledTasks() []ScheduledTask {
	return s.scheduler.GetTasks()
}

func (s *Service) RunTask(ctx context.Context, task ScheduledTask) error {
	for serverID, inst := range s.servers {
		if !inst.running || inst.RCON == nil {
			continue
		}
		var cmd string
		switch task.Type {
		case TaskRestart:
			cmd = "shutdown"
		case TaskBroadcast:
			cmd = fmt.Sprintf("say \"%s\"", task.Payload)
		case TaskBloodMoonCountdown:
			info, err := s.GetBloodMoonInfo(serverID)
			if err == nil {
				cmd = fmt.Sprintf("say \"距离血月还有 %d 天 %d 小时！\"", info.RemainingDays, info.RemainingHours)
			}
		case TaskCustomCommand:
			cmd = task.Payload
		}
		if cmd != "" {
			_ = inst.RCON.SendCommand(cmd)
		}
	}
	if s.emitter != nil {
		s.emitter.EmitEvent("dtd:task:ran", task)
	}
	return nil
}

func (s *Service) DeployToMachine(req RemoteDeployRequest) (*RemoteCommandResult, error) {
	if s.multiMachine == nil {
		return nil, fmt.Errorf("multi-machine bridge not configured")
	}
	output, err := s.multiMachine.ExecuteCommand(context.Background(), req.MachineID, "echo 'deploy placeholder'")
	if err != nil {
		return &RemoteCommandResult{Success: false, Output: output, Error: err.Error()}, nil
	}
	return &RemoteCommandResult{Success: true, Output: output}, nil
}

func (s *Service) StartRemoteServer(machineID string, installedDir string) (*RemoteCommandResult, error) {
	if s.multiMachine == nil {
		return nil, fmt.Errorf("multi-machine bridge not configured")
	}
	cmd := fmt.Sprintf("cd %s && ./startserver.sh -configfile=serverconfig.xml", installedDir)
	output, err := s.multiMachine.ExecuteCommand(context.Background(), machineID, cmd)
	if err != nil {
		return &RemoteCommandResult{Success: false, Output: output, Error: err.Error()}, nil
	}
	return &RemoteCommandResult{Success: true, Output: output}, nil
}

func (s *Service) handleRCONMessage(serverID string, line string) {
	s.mu.Lock()
	inst := s.servers[serverID]
	if inst == nil {
		s.mu.Unlock()
		return
	}

	inst.ConsoleBuf = append(inst.ConsoleBuf, ConsoleMessage{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "INFO",
		Message:   line,
	})
	if len(inst.ConsoleBuf) > 500 {
		inst.ConsoleBuf = inst.ConsoleBuf[len(inst.ConsoleBuf)-500:]
	}

	parser := NewPlayerListParser()
	players := parser.Parse(line)
	if len(players) > 0 {
		inst.Players = make(map[string]OnlinePlayer)
		for _, p := range players {
			inst.Players[p.ID] = p
		}
	}

	ev := s.logParser.ParseLine(line)
	if ev.Type != EventUnknown {
		switch ev.Type {
		case EventPlayerConnected:
			inst.Players[ev.PlayerName] = OnlinePlayer{
				ID:       ev.PlayerName,
				Name:     ev.PlayerName,
				Status:   "online",
				JoinedAt: ev.Timestamp.Format(time.RFC3339),
			}
		case EventPlayerDisconnected:
			delete(inst.Players, ev.PlayerName)
		case EventPlayerSpawned:
			if p, ok := inst.Players[ev.PlayerName]; ok {
				p.Status = "online"
				inst.Players[ev.PlayerName] = p
			}
		}
		inst.Events = append(inst.Events, PlayerEvent{
			Timestamp:  ev.Timestamp.Format(time.RFC3339),
			Type:       string(ev.Type),
			PlayerID:   ev.PlayerID,
			PlayerName: ev.PlayerName,
		})
	}
	s.mu.Unlock()

	if s.emitter != nil {
		s.emitter.EmitEvent("dtd:console:message", map[string]interface{}{
			"server_id": serverID,
			"line":      line,
		})
	}
}
