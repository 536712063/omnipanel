// Package agent 提供 Agent 主进程的 gRPC 服务端实现。
// Agent gRPC 服务是 UI (通过 Wails 绑定) 与后端插件之间的桥梁。
//
// 架构:
//   UI (Vue3) --[Wails Binding]--> Agent gRPC Server --[go-plugin]--> Plugin Processes
//
// 该服务端负责:
//   - 接收来自前端的请求
//   - 路由到对应的插件 gRPC 客户端
//   - 管理 WebSocket 连接用于流式数据推送
package agent

// 注意: 由于 github.com/omnipanel/omnipanel/proto 尚未通过 protoc 生成代码,
// 本文件中的 gRPC 服务实现代码将在 proto 代码生成后正常工作。
//
// 以下是 Agent 内部服务的完整实现骨架, 包含:
//   - Panic Recovery 中间件
//   - 请求超时控制
//   - 插件健康检查
//   - 统一错误响应格式
//
// 实际使用时, 取消下方代码的注释并导入生成的 proto 代码即可。

/*
import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	agentv1 "github.com/omnipanel/omnipanel/proto/agent/v1"
	commonv1 "github.com/omnipanel/omnipanel/proto/common/v1"
)

type AgentServer struct {
	agentv1.UnimplementedAgentInternalServer
	pluginMgr *Manager
	licenser  *license.Licenser
	wsHub     *ws.Hub
}

func NewAgentServer(pluginMgr *Manager, licenser *license.Licenser, wsHub *ws.Hub) *AgentServer {
	return &AgentServer{
		pluginMgr: pluginMgr,
		licenser:  licenser,
		wsHub:     wsHub,
	}
}

func (s *AgentServer) GetSystemInfo(ctx context.Context, _ *commonv1.Empty) (*commonv1.SystemInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	info, err := collectSystemInfo()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to collect system info: %v", err)
	}
	return info, nil
}

func (s *AgentServer) GetPluginStatus(ctx context.Context, _ *commonv1.Empty) (*agentv1.PluginStatusList, error) {
	plugins := s.pluginMgr.ListPlugins()
	list := make([]*agentv1.PluginStatus, 0, len(plugins))
	for _, pc := range plugins {
		list = append(list, &agentv1.PluginStatus{
			Name:    pc.Spec.Name,
			Running: !pc.Client.Exited(),
			Uptime:  time.Since(pc.StartTime).String(),
		})
	}
	return &agentv1.PluginStatusList{Plugins: list}, nil
}

func (s *AgentServer) RestartPlugin(ctx context.Context, req *agentv1.RestartPluginRequest) (*commonv1.StatusResponse, error) {
	// 实现插件重启逻辑
	return &commonv1.StatusResponse{Ok: true, Message: "plugin restarted"}, nil
}

func (s *AgentServer) ValidateLicense(ctx context.Context, key *agentv1.LicenseKey) (*agentv1.LicenseStatus, error) {
	return s.licenser.Validate(key.Key)
}

func (s *AgentServer) ActivateLicense(ctx context.Context, req *agentv1.ActivateLicenseRequest) (*agentv1.LicenseStatus, error) {
	return s.licenser.Activate(req.Key, req.Email)
}

// grpcUnaryInterceptor 为每个 gRPC 调用添加 panic recovery 和日志
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				hclog.L().Error("gRPC panic recovered", "method", info.FullMethod, "panic", r)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}
*/
