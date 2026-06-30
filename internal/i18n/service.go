package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

type ExtractedItem struct {
	Key         string `json:"key"`
	Original    string `json:"original"`
	Translation string `json:"translation"`
	Context     string `json:"context,omitempty"`
	Line        int    `json:"line"`
}

type ExtractRequest struct {
	SourceDir  string   `json:"source_dir"`
	Extensions []string `json:"extensions"`
}

type ExtractResult struct {
	TotalFiles int                          `json:"total_files"`
	TotalItems int                          `json:"total_items"`
	Files      map[string][]ExtractedItem   `json:"files"`
}

type Extractor struct {
	patterns []Pattern
}

type Pattern struct {
	Name  string
	Regex *regexp.Regexp
	Group int
}

type TranslationMap struct {
	Locale   string            `json:"locale"`
	Messages map[string]string `json:"messages"`
}

type Service struct {
	extractor *Extractor
}

func NewExtractor() *Extractor {
	return &Extractor{
		patterns: []Pattern{
			{Name: "vue_template", Regex: regexp.MustCompile(`[>\s]([\x{4e00}-\x{9fff}][^<"'\n]*[\x{4e00}-\x{9fff}])`), Group: 1},
			{Name: "string_literal", Regex: regexp.MustCompile(`['"]([\x{4e00}-\x{9fff}][^'"\n]*[\x{4e00}-\x{9fff}])['"]`), Group: 1},
		},
	}
}

func (e *Extractor) ExtractFromFile(path string) ([]ExtractedItem, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	var items []ExtractedItem
	seen := make(map[string]bool)

	for i, line := range lines {
		for _, pat := range e.patterns {
			matches := pat.Regex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > pat.Group {
					text := strings.TrimSpace(match[pat.Group])
					if text == "" || !containsChinese(text) || seen[text] {
						continue
					}
					seen[text] = true
					items = append(items, ExtractedItem{
						Key:      makeKey(text),
						Original: text,
						Line:     i + 1,
					})
				}
			}
		}
	}
	return items, nil
}

func (e *Extractor) ExtractFromDir(dir string, extensions []string) (map[string][]ExtractedItem, error) {
	result := make(map[string][]ExtractedItem)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		ext := filepath.Ext(path)
		for _, allowed := range extensions {
			if ext == allowed {
				items, err := e.ExtractFromFile(path)
				if err != nil {
					return err
				}
				if len(items) > 0 {
					result[path] = items
				}
				break
			}
		}
		return nil
	})
	return result, err
}

func containsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func makeKey(text string) string {
	runes := []rune(text)
	if len(runes) > 8 {
		runes = runes[:8]
	}
	return strings.ToLower(strings.ReplaceAll(string(runes), " ", "_"))
}

func BuildLocaleFile(items []ExtractedItem, locale string) TranslationMap {
	messages := make(map[string]string)
	for _, item := range items {
		messages[item.Key] = item.Translation
	}
	return TranslationMap{Locale: locale, Messages: messages}
}

func SaveLocaleFile(m TranslationMap, outputPath string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, data, 0644)
}

func ApplyTranslation(filePath string, items []ExtractedItem) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	content := string(data)
	for _, item := range items {
		replacement := fmt.Sprintf("{{ $t('%s') }}", item.Key)
		content = strings.Replace(content, item.Original, replacement, -1)
	}
	return os.WriteFile(filePath, []byte(content), 0644)
}

func NewService() *Service {
	return &Service{extractor: NewExtractor()}
}

func (s *Service) Extract(req ExtractRequest) (*ExtractResult, error) {
	if len(req.Extensions) == 0 {
		req.Extensions = []string{".vue", ".ts", ".js", ".json"}
	}
	files, err := s.extractor.ExtractFromDir(req.SourceDir, req.Extensions)
	if err != nil {
		return nil, err
	}
	totalItems := 0
	for _, items := range files {
		totalItems += len(items)
	}
	return &ExtractResult{
		TotalFiles: len(files),
		TotalItems: totalItems,
		Files:      files,
	}, nil
}

func (s *Service) GenerateLocaleFile(items []ExtractedItem, locale string, outputPath string) error {
	m := BuildLocaleFile(items, locale)
	return SaveLocaleFile(m, outputPath)
}

func (s *Service) PreviewTranslation(filePath string, items []ExtractedItem) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	content := string(data)
	for _, item := range items {
		replacement := fmt.Sprintf("{{ $t('%s') }}", item.Key)
		content = strings.Replace(content, item.Original, replacement, 1)
	}
	return content, nil
}

func (s *Service) ApplyTranslationFile(filePath string, items []ExtractedItem) error {
	return ApplyTranslation(filePath, items)
}

func (s *Service) BatchApplyTranslation(result ExtractResult) error {
	for filePath, items := range result.Files {
		if err := ApplyTranslation(filePath, items); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) SuggestKeys(sourceDir string, locale string) string {
	return filepath.Join(sourceDir, "locales", locale+".json")
}
