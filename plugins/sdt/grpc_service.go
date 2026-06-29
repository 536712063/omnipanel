package main

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-plugin"
	commonv1 "github.com/omnipanel/omnipanel/proto/common/v1"
	sdtv1 "github.com/omnipanel/omnipanel/proto/sdt/v1"
	"google.golang.org/grpc"
)

// SDTDGRPCPlugin 实现 hashicorp/go-plugin 的 gRPC 插件接口。
// 这是 SDTD 插件与 Agent 主进程之间的通信桥梁。
type SDTDGRPCPlugin struct {
	plugin.Plugin
	impl *SDTDPlugin
	ctx  context.Context
}

// GRPCServer 将 SDTD 的业务逻辑注册到 gRPC 服务器。
// 当 go-plugin 启动此插件的 gRPC 服务端时调用。
func (p *SDTDGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	sdtv1.RegisterSDTServiceServer(s, &SDTServiceServer{impl: p.impl})
	p.impl.logger.Info("SDTD gRPC server registered")
	return nil
}

// GRPCClient 返回 gRPC 客户端存根。
// 当 Agent 主进程需要调用此插件时调用。
func (p *SDTDGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return sdtv1.NewSDTServiceClient(c), nil
}

// ===========================================================================
// gRPC 服务实现 (SDTServiceServer)
// ===========================================================================

type SDTServiceServer struct {
	sdtv1.UnimplementedSDTServiceServer
	impl *SDTDPlugin
}

// ---------- 服务端生命周期 ----------

func (s *SDTServiceServer) InstallServer(req *sdtv1.InstallServerRequest, stream sdtv1.SDTService_InstallServerServer) error {
	startTime := time.Now()
	s.impl.logger.Info("开始安装七日杀服务端", "dir", req.InstallDir)

	outputCh, err := s.impl.installViaSteamCMD(stream.Context(), req.SteamUsername, req.SteamPassword)
	if err != nil {
		return fmt.Errorf("安装启动失败: %w", err)
	}

	for line := range outputCh {
		if err := stream.Send(&commonv1.LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Level:     "INFO",
			Message:   line,
		}); err != nil {
			return err
		}
	}

	s.impl.logger.Info("七日杀服务端安装完成",
		"duration", time.Since(startTime).String(),
	)
	return nil
}

func (s *SDTServiceServer) StartServer(ctx context.Context, req *commonv1.Empty) (*commonv1.StatusResponse, error) {
	if err := s.impl.startServer(ctx); err != nil {
		return &commonv1.StatusResponse{Ok: false, Message: err.Error()}, nil
	}
	return &commonv1.StatusResponse{Ok: true, Message: "服务器已启动"}, nil
}

func (s *SDTServiceServer) StopServer(ctx context.Context, req *commonv1.Empty) (*commonv1.StatusResponse, error) {
	if err := s.impl.stopServer(); err != nil {
		return &commonv1.StatusResponse{Ok: false, Message: err.Error()}, nil
	}
	return &commonv1.StatusResponse{Ok: true, Message: "服务器已停止"}, nil
}

func (s *SDTServiceServer) RestartServer(ctx context.Context, req *commonv1.Empty) (*commonv1.StatusResponse, error) {
	if err := s.impl.restartServer(ctx); err != nil {
		return &commonv1.StatusResponse{Ok: false, Message: err.Error()}, nil
	}
	return &commonv1.StatusResponse{Ok: true, Message: "服务器已重启"}, nil
}

func (s *SDTServiceServer) GetServerStatus(ctx context.Context, req *commonv1.Empty) (*sdtv1.ServerStatusResponse, error) {
	return &sdtv1.ServerStatusResponse{
		Installed:     true,
		Running:       s.impl.isServerRunning(),
		OnlinePlayers: 5,
		MaxPlayers:    16,
		Tps:           19.2,
	}, nil
}

// ---------- 配置文件 ----------

func (s *SDTServiceServer) GetServerConfig(ctx context.Context, req *commonv1.Empty) (*sdtv1.ServerConfig, error) {
	cfg, err := s.impl.readServerConfig()
	if err != nil {
		return nil, err
	}
	return &sdtv1.ServerConfig{
		ServerName:          cfg.ServerName,
		ServerDescription:   cfg.ServerDescription,
		ServerPassword:      cfg.ServerPassword,
		ServerMaxPlayerCount: int32(cfg.ServerMaxPlayerCount),
		ServerPort:          int32(cfg.ServerPort),
		GameDifficulty:      int32(cfg.GameDifficulty),
		GameMode:            cfg.GameMode,
		WorldGenSeed:        cfg.WorldGenSeed,
		WorldGenSize:        int32(cfg.WorldGenSize),
		MaxSpawnedZombies:   int32(cfg.MaxSpawnedZombies),
		MaxSpawnedAnimals:   int32(cfg.MaxSpawnedAnimals),
		EacEnabled:          cfg.EACEnabled,
		BloodMoonFrequency:  int32(cfg.BloodMoonFrequency),
		DropOnDeath:         int32(cfg.DropOnDeath),
		DropOnQuit:          int32(cfg.DropOnQuit),
		TelnetEnabled:       cfg.TelnetEnabled,
		TelnetPort:          int32(cfg.TelnetPort),
	}, nil
}

func (s *SDTServiceServer) UpdateServerConfig(ctx context.Context, req *sdtv1.UpdateServerConfigRequest) (*commonv1.StatusResponse, error) {
	cfg := &ServerConfigXML{
		ServerName:          req.Config.ServerName,
		ServerDescription:   req.Config.ServerDescription,
		ServerPassword:      req.Config.ServerPassword,
		ServerMaxPlayerCount: int(req.Config.ServerMaxPlayerCount),
		ServerPort:          int(req.Config.ServerPort),
		GameDifficulty:      int(req.Config.GameDifficulty),
		GameMode:            req.Config.GameMode,
		WorldGenSeed:        req.Config.WorldGenSeed,
		WorldGenSize:        int(req.Config.WorldGenSize),
		MaxSpawnedZombies:   int(req.Config.MaxSpawnedZombies),
		MaxSpawnedAnimals:   int(req.Config.MaxSpawnedAnimals),
		EACEnabled:          req.Config.EacEnabled,
		BloodMoonFrequency:  int(req.Config.BloodMoonFrequency),
		DropOnDeath:         int(req.Config.DropOnDeath),
		DropOnQuit:          int(req.Config.DropOnQuit),
		TelnetEnabled:       req.Config.TelnetEnabled,
		TelnetPort:          int(req.Config.TelnetPort),
	}

	errors, warnings := s.impl.validateConfig(cfg)
	if len(errors) > 0 {
		msg := "配置验证失败: "
		for _, e := range errors {
			msg += e + "; "
		}
		return &commonv1.StatusResponse{Ok: false, Message: msg}, nil
	}

	if len(warnings) > 0 {
		s.impl.logger.Warn("配置包含警告", "warnings", warnings)
	}

	if err := s.impl.writeServerConfig(cfg); err != nil {
		return &commonv1.StatusResponse{Ok: false, Message: err.Error()}, nil
	}

	return &commonv1.StatusResponse{Ok: true, Message: "配置已保存"}, nil
}

func (s *SDTServiceServer) ValidateConfig(ctx context.Context, req *commonv1.Empty) (*sdtv1.ValidationResult, error) {
	cfg, err := s.impl.readServerConfig()
	if err != nil {
		return &sdtv1.ValidationResult{Valid: false, Errors: []string{err.Error()}}, nil
	}

	errors, warnings := s.impl.validateConfig(cfg)
	return &sdtv1.ValidationResult{
		Valid:    len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}, nil
}

// ---------- 控制台 ----------

func (s *SDTServiceServer) SendConsoleCommand(ctx context.Context, req *sdtv1.ConsoleCommandRequest) (*sdtv1.ConsoleCommandResponse, error) {
	output, err := s.impl.sendCommand(req.Command)
	if err != nil {
		return &sdtv1.ConsoleCommandResponse{Output: err.Error(), Success: false}, nil
	}
	return &sdtv1.ConsoleCommandResponse{Output: output, Success: true}, nil
}

// ---------- 存档备份 ----------

func (s *SDTServiceServer) CreateBackup(ctx context.Context, req *commonv1.Empty) (*sdtv1.BackupResponse, error) {
	name, err := s.impl.createBackup()
	if err != nil {
		return nil, err
	}
	return &sdtv1.BackupResponse{
		Id:        name,
		Name:      name,
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *SDTServiceServer) RestoreBackup(ctx context.Context, req *sdtv1.RestoreBackupRequest) (*commonv1.StatusResponse, error) {
	if err := s.impl.restoreBackup(req.BackupId); err != nil {
		return &commonv1.StatusResponse{Ok: false, Message: err.Error()}, nil
	}
	return &commonv1.StatusResponse{Ok: true, Message: "存档已恢复"}, nil
}
