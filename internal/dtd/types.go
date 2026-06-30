package dtd

import "context"

type ServerConfig struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Host         string `json:"host"`
	GamePort     int    `json:"game_port"`
	TelnetPort   int    `json:"telnet_port"`
	TelnetPass   string `json:"telnet_pass"`
	RCONPort     int    `json:"rcon_port"`
	RCONPass     string `json:"rcon_pass"`
	SaveDir      string `json:"save_dir"`
	LogPath      string `json:"log_path"`
	BloodMoonDay int    `json:"blood_moon_day"`
	InstalledDir string `json:"installed_dir"`
}

type OnlinePlayer struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	PlatformID string    `json:"platform_id,omitempty"`
	Status     string    `json:"status"`
	JoinedAt   string    `json:"joined_at"`
	Position   *Position `json:"position,omitempty"`
}

type PlayerEvent struct {
	Timestamp  string `json:"timestamp"`
	Type       string `json:"type"`
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
}

type ConsoleMessage struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

type EventsEmitter interface {
	EmitEvent(eventName string, optionalData ...interface{})
}

type MultiMachineBridge interface {
	ExecuteCommand(ctx context.Context, machineID string, command string) (string, error)
	UploadToMachine(ctx context.Context, machineID string, localPath string, remotePath string) error
	DownloadFromMachine(ctx context.Context, machineID string, remotePath string, localPath string) error
}

type BloodMoonInfo struct {
	CurrentDay    int `json:"current_day"`
	NextBloodMoon int `json:"next_blood_moon"`
	RemainingDays int `json:"remaining_days"`
	RemainingHours int `json:"remaining_hours"`
}

type RemoteCommandResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

type RemoteDeployRequest struct {
	MachineID    string       `json:"machine_id"`
	InstallDir   string       `json:"install_dir"`
	ServerConfig ServerConfig `json:"server_config"`
}
