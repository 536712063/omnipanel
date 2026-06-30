package browser

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type Service struct{}

type BrowserInfo struct {
	UserAgent  string `json:"user_agent"`
	Platform   string `json:"platform"`
	IsEmbedded bool   `json:"is_embedded"`
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) OpenExternalURL(ctx context.Context, url string) error {
	if err := validateURL(url); err != nil {
		return err
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

func (s *Service) GetInfo() BrowserInfo {
	return BrowserInfo{
		UserAgent:  "OmniPanel/1.0",
		Platform:   runtime.GOOS,
		IsEmbedded: true,
	}
}

func (s *Service) ValidateURL(url string) error {
	return validateURL(url)
}

func validateURL(url string) error {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("url must start with http:// or https://")
	}
	return nil
}
