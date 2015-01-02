package lib

import (
	"fmt"
	"path/filepath"
)

func GetImageMimeType(filename string) (string, error) {
	switch filepath.Ext(filename) {
	case ".gif":
		return "image/gif", nil
	case ".svg":
		return "image/svg+xml", nil
	default:
		return "", fmt.Errorf("Unsupported image file extension: %s. Supported types are .gif and .svg", filepath.Ext(filename))
	}
}
