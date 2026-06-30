package cloud

import "fmt"

type OAuthProvider struct {
	config ProviderConfig
}

func NewOAuthProvider(cfg ProviderConfig) *OAuthProvider {
	return &OAuthProvider{config: cfg}
}

func (p *OAuthProvider) Name() string { return p.config.Name }
func (p *OAuthProvider) List(ctx Context, path string) ([]FileInfo, error) {
	return nil, fmt.Errorf("not yet implemented")
}
func (p *OAuthProvider) Upload(ctx Context, localPath string, remotePath string, progressCh chan<- ProgressEvent) error {
	return fmt.Errorf("not yet implemented")
}
func (p *OAuthProvider) Download(ctx Context, remotePath string, localPath string, progressCh chan<- ProgressEvent) error {
	return fmt.Errorf("not yet implemented")
}
func (p *OAuthProvider) Copy(ctx Context, srcPath string, dstPath string) error {
	return fmt.Errorf("not yet implemented")
}
func (p *OAuthProvider) Move(ctx Context, srcPath string, dstPath string) error {
	return fmt.Errorf("not yet implemented")
}
func (p *OAuthProvider) Rename(ctx Context, path string, newName string) error {
	return fmt.Errorf("not yet implemented")
}
func (p *OAuthProvider) Delete(ctx Context, path string) error {
	return fmt.Errorf("not yet implemented")
}
func (p *OAuthProvider) Mkdir(ctx Context, path string) error {
	return fmt.Errorf("not yet implemented")
}
func (p *OAuthProvider) GetPreview(ctx Context, path string) (PreviewInfo, error) {
	return PreviewInfo{}, fmt.Errorf("not yet implemented")
}
func (p *OAuthProvider) OAuthURL() (string, error) {
	switch p.config.Type {
	case "baidu":
		return fmt.Sprintf("https://openapi.baidu.com/oauth/2.0/authorize?client_id=%s&redirect_uri=%s&response_type=code", p.config.ClientID, p.config.RedirectURI), nil
	case "aliyun":
		return fmt.Sprintf("https://auth.aliyundrive.com/v2/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code", p.config.ClientID, p.config.RedirectURI), nil
	}
	return "", fmt.Errorf("OAuth not available for %s", p.config.Type)
}
func (p *OAuthProvider) HandleOAuthCallback(code string, state string) (*OAuthToken, error) {
	return nil, fmt.Errorf("not yet implemented")
}

type BaiduOAuthProvider struct{ *OAuthProvider }
type AliyunOAuthProvider struct{ *OAuthProvider }
type QuarkOAuthProvider struct{ *OAuthProvider }
type P115OAuthProvider struct{ *OAuthProvider }

func NewBaiduOAuthProvider(cfg ProviderConfig) Provider {
	return &BaiduOAuthProvider{OAuthProvider: NewOAuthProvider(cfg)}
}
func NewAliyunOAuthProvider(cfg ProviderConfig) Provider {
	return &AliyunOAuthProvider{OAuthProvider: NewOAuthProvider(cfg)}
}
func NewQuarkOAuthProvider(cfg ProviderConfig) Provider {
	return &QuarkOAuthProvider{OAuthProvider: NewOAuthProvider(cfg)}
}
func NewP115OAuthProvider(cfg ProviderConfig) Provider {
	return &P115OAuthProvider{OAuthProvider: NewOAuthProvider(cfg)}
}
