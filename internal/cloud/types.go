package cloud

import "time"

type Provider interface {
	Name() string
	List(ctx Context, path string) ([]FileInfo, error)
	Upload(ctx Context, localPath string, remotePath string, progressCh chan<- ProgressEvent) error
	Download(ctx Context, remotePath string, localPath string, progressCh chan<- ProgressEvent) error
	Copy(ctx Context, srcPath string, dstPath string) error
	Move(ctx Context, srcPath string, dstPath string) error
	Rename(ctx Context, path string, newName string) error
	Delete(ctx Context, path string) error
	Mkdir(ctx Context, path string) error
	GetPreview(ctx Context, path string) (PreviewInfo, error)
	OAuthURL() (string, error)
	HandleOAuthCallback(code string, state string) (*OAuthToken, error)
}

type Context struct {
	ContextID string
}

type FileInfo struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	IsDir        bool      `json:"is_dir"`
	ModifiedAt   time.Time `json:"modified_at"`
	ThumbnailURL string    `json:"thumbnail_url"`
	MimeType     string    `json:"mime_type"`
}

type ProgressEvent struct {
	Type        string  `json:"type"`
	ContextID   string  `json:"context_id"`
	Path        string  `json:"path"`
	Transferred int64   `json:"transferred"`
	Total       int64   `json:"total"`
	Speed       float64 `json:"speed"`
	Error       string  `json:"error,omitempty"`
}

type PreviewInfo struct {
	Type      string `json:"type"`
	Data      []byte `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
	MimeType  string `json:"mime_type"`
	Extension string `json:"extension"`
	Size      int64  `json:"size"`
}

type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	Provider     string    `json:"provider"`
}

type ProviderConfig struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	BaseURL      string `json:"base_url"`
	Token        string `json:"token"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	OAuthToken   *OAuthToken
}

type MultiMachineSyncRequest struct {
	MachineID    string `json:"machine_id"`
	RemoteFile   string `json:"remote_file"`
	CloudPath    string `json:"cloud_path"`
	ProviderName string `json:"provider_name"`
}
