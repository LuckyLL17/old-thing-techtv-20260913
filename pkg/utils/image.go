package utils

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

func ResizeImage(srcPath, dstPath string, maxWidth, maxHeight int) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()
	img, format, err := image.Decode(f)
	if err != nil {
		return err
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= maxWidth && h <= maxHeight {
		return nil
	}
	ratio := float64(maxWidth) / float64(w)
	if float64(maxHeight)/float64(h) < ratio {
		ratio = float64(maxHeight) / float64(h)
	}
	newW, newH := int(float64(w)*ratio), int(float64(h)*ratio)
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return err
	}
	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return jpeg.Encode(out, dst, &jpeg.Options{Quality: 85})
	case "png":
		return png.Encode(out, dst)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func GenerateThumbnail(srcPath string, size int) (string, error) {
	ext := filepath.Ext(srcPath)
	base := strings.TrimSuffix(srcPath, ext)
	thumbPath := fmt.Sprintf("%s_thumb%d%s", base, size, ext)
	if err := ResizeImage(srcPath, thumbPath, size, size); err != nil {
		return "", err
	}
	return thumbPath, nil
}
