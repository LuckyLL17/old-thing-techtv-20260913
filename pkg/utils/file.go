package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var allowedImageExt = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

func SaveUploadedFile(file *multipart.FileHeader, destDir string, maxSize int64) (string, error) {
	if file.Size > maxSize {
		return "", fmt.Errorf("file size exceeds limit %d bytes", maxSize)
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExt[ext] {
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}
	h := md5.New()
	io.WriteString(h, file.Filename+time.Now().String())
	name := hex.EncodeToString(h.Sum(nil)) + ext
	fullPath := filepath.Join(destDir, name)
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}
	return fullPath, nil
}

func RemoveFile(path string) error {
	if path == "" {
		return nil
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}
