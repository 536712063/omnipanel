package cloud

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type Service struct {
	providers    map[string]Provider
	providersMu  sync.RWMutex
	emitter      EventsEmitter
	multiMachine MultiMachineBridge
}

type EventsEmitter interface {
	EmitEvent(eventName string, optionalData ...interface{})
}

type MultiMachineBridge interface {
	DownloadFromMachine(ctx context.Context, machineID string, remotePath string, localPath string) error
	UploadToMachine(ctx context.Context, machineID string, localPath string, remotePath string) error
}

func NewService(emitter EventsEmitter, bridge MultiMachineBridge) *Service {
	return &Service{
		providers:    make(map[string]Provider),
		emitter:      emitter,
		multiMachine: bridge,
	}
}

func (s *Service) RegisterProvider(name string, provider Provider) {
	s.providersMu.Lock()
	defer s.providersMu.Unlock()
	s.providers[name] = provider
}

func (s *Service) GetProvider(name string) (Provider, error) {
	s.providersMu.RLock()
	defer s.providersMu.RUnlock()
	p, ok := s.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return p, nil
}

func (s *Service) ListProviders() []string {
	s.providersMu.RLock()
	defer s.providersMu.RUnlock()
	names := make([]string, 0, len(s.providers))
	for name := range s.providers {
		names = append(names, name)
	}
	return names
}

func (s *Service) AddProviderFromConfig(cfg ProviderConfig) error {
	var provider Provider
	switch cfg.Type {
	case "alist", "openlist":
		provider = NewAlistProvider(cfg)
	case "baidu":
		provider = NewBaiduOAuthProvider(cfg)
	case "aliyun":
		provider = NewAliyunOAuthProvider(cfg)
	case "quark":
		provider = NewQuarkOAuthProvider(cfg)
	case "115":
		provider = NewP115OAuthProvider(cfg)
	default:
		return fmt.Errorf("unsupported provider type: %s", cfg.Type)
	}
	s.RegisterProvider(cfg.Name, provider)
	return nil
}

func (s *Service) ListFiles(providerName string, path string) ([]FileInfo, error) {
	p, err := s.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	ctx := Context{ContextID: uuid.New().String()}
	return p.List(ctx, path)
}

func (s *Service) CopyFile(providerName, srcPath, dstPath string) error {
	p, err := s.GetProvider(providerName)
	if err != nil {
		return err
	}
	ctx := Context{ContextID: uuid.New().String()}
	return p.Copy(ctx, srcPath, dstPath)
}

func (s *Service) MoveFile(providerName, srcPath, dstPath string) error {
	p, err := s.GetProvider(providerName)
	if err != nil {
		return err
	}
	ctx := Context{ContextID: uuid.New().String()}
	return p.Move(ctx, srcPath, dstPath)
}

func (s *Service) RenameFile(providerName, path, newName string) error {
	p, err := s.GetProvider(providerName)
	if err != nil {
		return err
	}
	ctx := Context{ContextID: uuid.New().String()}
	return p.Rename(ctx, path, newName)
}

func (s *Service) DeleteFile(providerName, path string) error {
	p, err := s.GetProvider(providerName)
	if err != nil {
		return err
	}
	ctx := Context{ContextID: uuid.New().String()}
	return p.Delete(ctx, path)
}

func (s *Service) Mkdir(providerName, path string) error {
	p, err := s.GetProvider(providerName)
	if err != nil {
		return err
	}
	ctx := Context{ContextID: uuid.New().String()}
	return p.Mkdir(ctx, path)
}

func (s *Service) UploadFile(providerName, localPath, remotePath string) (string, error) {
	p, err := s.GetProvider(providerName)
	if err != nil {
		return "", err
	}
	ctx := Context{ContextID: uuid.New().String()}
	progressCh := make(chan ProgressEvent, 128)

	go func() {
		defer close(progressCh)
		_ = p.Upload(ctx, localPath, remotePath, progressCh)
	}()

	go s.forwardProgress(ctx.ContextID, progressCh)
	return ctx.ContextID, nil
}

func (s *Service) DownloadFile(providerName, remotePath, localPath string) (string, error) {
	p, err := s.GetProvider(providerName)
	if err != nil {
		return "", err
	}
	ctx := Context{ContextID: uuid.New().String()}
	progressCh := make(chan ProgressEvent, 128)

	go func() {
		defer close(progressCh)
		_ = p.Download(ctx, remotePath, localPath, progressCh)
	}()

	go s.forwardProgress(ctx.ContextID, progressCh)
	return ctx.ContextID, nil
}

func (s *Service) GetPreview(providerName, path string) (PreviewInfo, error) {
	p, err := s.GetProvider(providerName)
	if err != nil {
		return PreviewInfo{}, err
	}
	ctx := Context{ContextID: uuid.New().String()}
	return p.GetPreview(ctx, path)
}

func (s *Service) OAuthURL(providerName string) (string, error) {
	p, err := s.GetProvider(providerName)
	if err != nil {
		return "", err
	}
	return p.OAuthURL()
}

func (s *Service) HandleOAuthCallback(providerName, code, state string) (*OAuthToken, error) {
	p, err := s.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	return p.HandleOAuthCallback(code, state)
}

func (s *Service) SyncFromMachine(req MultiMachineSyncRequest) (string, error) {
	if s.multiMachine == nil {
		return "", fmt.Errorf("multi-machine bridge not configured")
	}
	p, err := s.GetProvider(req.ProviderName)
	if err != nil {
		return "", err
	}

	ctxID := uuid.New().String()
	localTemp := fmt.Sprintf("/tmp/omnipanel/machine-sync-%s", ctxID)

	progressCh := make(chan ProgressEvent, 128)
	go s.forwardProgress(ctxID, progressCh)

	go func() {
		defer close(progressCh)
		if err := s.multiMachine.DownloadFromMachine(context.Background(), req.MachineID, req.RemoteFile, localTemp); err != nil {
			progressCh <- ProgressEvent{Type: "error", ContextID: ctxID, Error: err.Error()}
			return
		}
		ctx := Context{ContextID: ctxID}
		_ = p.Upload(ctx, localTemp, req.CloudPath, progressCh)
	}()

	return ctxID, nil
}

func (s *Service) forwardProgress(contextID string, progressCh <-chan ProgressEvent) {
	for ev := range progressCh {
		if s.emitter != nil {
			s.emitter.EmitEvent("cloud:progress:event", ev)
		}
	}
}
