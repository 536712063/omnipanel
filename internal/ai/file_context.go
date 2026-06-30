package ai

import (
	"encoding/base64"
	"fmt"
	"mime"
	"path/filepath"
)

type FileContextHandler struct {
	maxImageSize int64
	maxFileSize  int64
}

func NewFileContextHandler() *FileContextHandler {
	return &FileContextHandler{
		maxImageSize: 10 * 1024 * 1024,
		maxFileSize:  20 * 1024 * 1024,
	}
}

type SupportedFileResult struct {
	Part      ContentPart
	IsImage   bool
	Thumbnail string
}

func (h *FileContextHandler) ProcessFile(name string, data []byte) (SupportedFileResult, error) {
	if len(data) == 0 {
		return SupportedFileResult{}, fmt.Errorf("empty file")
	}

	contentType := mime.TypeByExtension(filepath.Ext(name))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	isImage := false
	switch filepath.Ext(name) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp":
		isImage = true
	}

	if isImage && int64(len(data)) > h.maxImageSize {
		return SupportedFileResult{}, fmt.Errorf("image exceeds max size %d", h.maxImageSize)
	}
	if !isImage && int64(len(data)) > h.maxFileSize {
		return SupportedFileResult{}, fmt.Errorf("file exceeds max size %d", h.maxFileSize)
	}

	part := ContentPart{
		Type:     "file",
		FileName: name,
		FileType: contentType,
		FileData: data,
	}
	if isImage {
		part.Type = "image_url"
		part.ImageURL = fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(data))
	}

	return SupportedFileResult{
		Part:    part,
		IsImage: isImage,
	}, nil
}
