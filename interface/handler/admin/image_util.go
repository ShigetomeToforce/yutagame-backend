package admin

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const maxImageSizeBytes = 5 * 1024 * 1024

var (
	errImageRequired   = errors.New("画像ファイルを選択してください")
	errImageTooLarge   = errors.New("画像サイズは5MB以下にしてください")
	errImageTypeDenied = errors.New("対応していない画像形式です（jpg/png/webpのみ）")
	nonCodeCharRegexp  = regexp.MustCompile(`[^A-Za-z0-9_-]`)
)

func saveUploadedImage(fileHeader *multipart.FileHeader, category string, code string) (string, error) {
	if fileHeader == nil {
		return "", errImageRequired
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, maxImageSizeBytes+1))
	if err != nil {
		return "", err
	}
	if len(content) > maxImageSizeBytes {
		return "", errImageTooLarge
	}

	contentType := http.DetectContentType(content)
	ext, ok := extensionFromContentType(contentType)
	if !ok {
		return "", errImageTypeDenied
	}

	safeCode := sanitizeCode(code)
	if safeCode == "" {
		safeCode = "image"
	}

	fileName := fmt.Sprintf("%s_%d%s", safeCode, time.Now().UnixNano(), ext)
	imageKey := filepath.ToSlash(filepath.Join(category, fileName))
	fullPath := filepath.Join("storage", filepath.FromSlash(imageKey))

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(fullPath, content, 0o644); err != nil {
		return "", err
	}

	return imageKey, nil
}

func deleteStoredImage(imageKey string) error {
	if strings.TrimSpace(imageKey) == "" {
		return nil
	}

	fullPath := filepath.Join("storage", filepath.FromSlash(imageKey))
	err := os.Remove(fullPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func extensionFromContentType(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}

func sanitizeCode(code string) string {
	trimmed := strings.TrimSpace(code)
	return nonCodeCharRegexp.ReplaceAllString(trimmed, "-")
}
