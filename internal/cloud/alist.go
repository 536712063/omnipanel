package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type AlistProvider struct {
	config ProviderConfig
	client *http.Client
}

func NewAlistProvider(cfg ProviderConfig) *AlistProvider {
	return &AlistProvider{
		config: cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *AlistProvider) Name() string {
	return p.config.Name
}

func (p *AlistProvider) apiRequest(method string, endpoint string, body interface{}, result interface{}) error {
	url := strings.TrimRight(p.config.BaseURL, "/") + endpoint
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.config.Token != "" {
		req.Header.Set("Authorization", p.config.Token)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("alist api error: status %d", resp.StatusCode)
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

type alistListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Content []struct {
			Name     string `json:"name"`
			Path     string `json:"path"`
			Size     int64  `json:"size"`
			IsDir    bool   `json:"is_dir"`
			Modified string `json:"modified"`
			Thumb    string `json:"thumb"`
			Type     int    `json:"type"`
		} `json:"content"`
	} `json:"data"`
}

func (p *AlistProvider) List(ctx Context, dirPath string) ([]FileInfo, error) {
	var resp alistListResponse
	err := p.apiRequest("POST", "/api/fs/list", map[string]interface{}{
		"path":     dirPath,
		"page":     1,
		"per_page": 100,
	}, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, fmt.Errorf("alist list error: %s", resp.Message)
	}

	files := make([]FileInfo, 0, len(resp.Data.Content))
	for _, item := range resp.Data.Content {
		modTime, _ := time.Parse(time.RFC3339, item.Modified)
		files = append(files, FileInfo{
			Name:         item.Name,
			Path:         path.Join(dirPath, item.Name),
			Size:         item.Size,
			IsDir:        item.IsDir,
			ModifiedAt:   modTime,
			ThumbnailURL: item.Thumb,
		})
	}
	return files, nil
}

func (p *AlistProvider) Copy(ctx Context, srcPath string, dstPath string) error {
	return p.apiRequest("POST", "/api/fs/copy", map[string]interface{}{
		"src_dir": filepath.Dir(srcPath),
		"dst_dir": filepath.Dir(dstPath),
		"names":   []string{filepath.Base(srcPath)},
	}, nil)
}

func (p *AlistProvider) Move(ctx Context, srcPath string, dstPath string) error {
	return p.apiRequest("POST", "/api/fs/move", map[string]interface{}{
		"src_dir": filepath.Dir(srcPath),
		"dst_dir": filepath.Dir(dstPath),
		"names":   []string{filepath.Base(srcPath)},
	}, nil)
}

func (p *AlistProvider) Rename(ctx Context, filePath string, newName string) error {
	return p.apiRequest("POST", "/api/fs/rename", map[string]interface{}{
		"path":     filePath,
		"new_name": newName,
	}, nil)
}

func (p *AlistProvider) Delete(ctx Context, filePath string) error {
	return p.apiRequest("POST", "/api/fs/remove", map[string]interface{}{
		"dir":   filepath.Dir(filePath),
		"names": []string{filepath.Base(filePath)},
	}, nil)
}

func (p *AlistProvider) Mkdir(ctx Context, dirPath string) error {
	return p.apiRequest("POST", "/api/fs/mkdir", map[string]interface{}{
		"path": dirPath,
	}, nil)
}

func (p *AlistProvider) Upload(ctx Context, localPath string, remotePath string, progressCh chan<- ProgressEvent) error {
	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}
	total := stat.Size()

	pr := &progressReader{
		Reader:      f,
		total:       total,
		contextID:   ctx.ContextID,
		path:        remotePath,
		eventType:   "upload",
		progressCh:  progressCh,
		start:       time.Now(),
	}

	var uploadResp struct {
		Code int `json:"code"`
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	err = p.apiRequest("PUT", "/api/fs/put", map[string]interface{}{
		"path": remotePath,
	}, &uploadResp)
	if err != nil {
		return err
	}
	if uploadResp.Code != 200 {
		return fmt.Errorf("failed to get upload url")
	}

	req, err := http.NewRequest("PUT", uploadResp.Data.URL, pr)
	if err != nil {
		return err
	}
	req.ContentLength = total
	resp, err := p.client.Do(req)
	if err != nil {
		pr.emitError(err)
		return err
	}
	defer resp.Body.Close()

	if progressCh != nil {
		progressCh <- ProgressEvent{Type: "done", ContextID: ctx.ContextID, Path: remotePath, Total: total, Transferred: total}
	}
	return nil
}

func (p *AlistProvider) Download(ctx Context, remotePath string, localPath string, progressCh chan<- ProgressEvent) error {
	var downloadResp struct {
		Code int `json:"code"`
		Data struct {
			URL  string `json:"url"`
			Size int64  `json:"size"`
		} `json:"data"`
	}
	err := p.apiRequest("POST", "/api/fs/get", map[string]interface{}{
		"path": remotePath,
	}, &downloadResp)
	if err != nil {
		return err
	}
	if downloadResp.Code != 200 {
		return fmt.Errorf("failed to get download url")
	}

	resp, err := p.client.Get(downloadResp.Data.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer out.Close()

	total := downloadResp.Data.Size
	if total == 0 {
		total = resp.ContentLength
	}
	pr := &progressReader{
		Reader:      resp.Body,
		total:       total,
		contextID:   ctx.ContextID,
		path:        remotePath,
		eventType:   "download",
		progressCh:  progressCh,
		start:       time.Now(),
	}

	_, err = io.Copy(out, pr)
	if err != nil {
		pr.emitError(err)
		return err
	}

	if progressCh != nil {
		progressCh <- ProgressEvent{Type: "done", ContextID: ctx.ContextID, Path: remotePath, Total: total, Transferred: total}
	}
	return nil
}

func (p *AlistProvider) GetPreview(ctx Context, filePath string) (PreviewInfo, error) {
	var previewResp struct {
		Code int `json:"code"`
		Data struct {
			URL    string `json:"url"`
			RawURL string `json:"raw_url"`
			Size   int64  `json:"size"`
		} `json:"data"`
	}
	err := p.apiRequest("POST", "/api/fs/get", map[string]interface{}{
		"path": filePath,
	}, &previewResp)
	if err != nil {
		return PreviewInfo{}, err
	}
	if previewResp.Code != 200 {
		return PreviewInfo{}, fmt.Errorf("failed to get preview")
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	info := PreviewInfo{
		URL:       previewResp.Data.RawURL,
		Size:      previewResp.Data.Size,
		Extension: ext,
	}

	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp":
		info.Type = "image"
		info.MimeType = "image/" + strings.TrimPrefix(ext, ".")
	case ".mp4", ".mkv", ".avi", ".mov", ".webm":
		info.Type = "video"
		info.MimeType = "video/mp4"
	case ".mp3", ".flac", ".wav", ".aac", ".ogg":
		info.Type = "audio"
		info.MimeType = "audio/mpeg"
	case ".txt", ".md", ".json", ".yaml", ".yml", ".go", ".ts", ".js", ".vue":
		info.Type = "text"
		info.MimeType = "text/plain"
	default:
		info.Type = "unsupported"
	}

	return info, nil
}

func (p *AlistProvider) OAuthURL() (string, error) {
	return "", fmt.Errorf("Alist does not require OAuth")
}

func (p *AlistProvider) HandleOAuthCallback(code string, state string) (*OAuthToken, error) {
	return nil, fmt.Errorf("Alist does not require OAuth")
}

type progressReader struct {
	io.Reader
	total       int64
	transferred int64
	contextID   string
	path        string
	eventType   string
	progressCh  chan<- ProgressEvent
	start       time.Time
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	pr.transferred += int64(n)

	elapsed := time.Since(pr.start).Seconds()
	var speed float64
	if elapsed > 0 {
		speed = float64(pr.transferred) / elapsed
	}

	if pr.progressCh != nil {
		pr.progressCh <- ProgressEvent{
			Type:        pr.eventType,
			ContextID:   pr.contextID,
			Path:        pr.path,
			Transferred: pr.transferred,
			Total:       pr.total,
			Speed:       speed,
		}
	}
	return n, err
}

func (pr *progressReader) emitError(err error) {
	if pr.progressCh != nil {
		pr.progressCh <- ProgressEvent{
			Type:      "error",
			ContextID: pr.contextID,
			Path:      pr.path,
			Error:     err.Error(),
		}
	}
}
