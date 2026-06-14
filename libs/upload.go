package libs

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// UploadDir is the directory (relative to the application working directory)
// where uploaded images are stored and served from.
const UploadDir = "static/upload"

// EnsureDir makes sure dir (and any missing parent directories) exists.
// It is safe to call when the directory is already present.
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, os.ModePerm)
}

// SanitizeFileName strips any path information and replaces characters that are
// unsafe for a filename with an underscore. Leading/trailing dots are removed so
// the result can never become a hidden file or a parent-directory reference
// ("." / ".."). An empty result is returned as-is so callers can fall back to a
// default base name.
func SanitizeFileName(name string) string {
	// Drop any directory component first to defend against path traversal.
	name = path.Base(filepath.ToSlash(strings.TrimSpace(name)))

	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return strings.Trim(b.String(), ".")
}

// UniqueUploadPath ensures dir exists and returns a web-style (forward slash)
// path inside dir that does not yet exist on disk. Because the returned name is
// always collision-free, uploading two files with the same original name never
// overwrites a previously stored file.
func UniqueUploadPath(dir, originalName string) (string, error) {
	if err := EnsureDir(dir); err != nil {
		return "", err
	}

	clean := SanitizeFileName(originalName)
	ext := path.Ext(clean)
	base := strings.TrimSuffix(clean, ext)
	if base == "" {
		base = "upload"
	}

	for i := 0; i < 1000; i++ {
		name := fmt.Sprintf("%s_%d_%d_%s%s", base, time.Now().UnixNano(), i, GetRandomString(8), ext)
		candidate := dir + "/" + name
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not allocate a unique upload name for %q", originalName)
}

// SaveUploaded stores the bytes read from src under dir using a collision-free
// name derived from originalName and returns the stored web-style path. The
// upload directory is created automatically when missing. If writing fails the
// partially written file is removed so a failed upload never leaves a corrupt
// image (and, because the name is unique, it can never clobber an existing one).
func SaveUploaded(dir, originalName string, src io.Reader) (string, error) {
	savePath, err := UniqueUploadPath(dir, originalName)
	if err != nil {
		return "", err
	}

	dst, err := os.Create(savePath)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		os.Remove(savePath)
		return "", err
	}
	if err := dst.Close(); err != nil {
		os.Remove(savePath)
		return "", err
	}
	return savePath, nil
}

// ResolveImageURL chooses which image path a record should keep after an update.
// When a new image was uploaded it returns newPath; otherwise it preserves
// oldPath. This keeps the "user did not reselect an image" and "user replaced the
// image" cases consistent across the article and banner update flows.
func ResolveImageURL(oldPath, newPath string, uploaded bool) string {
	if uploaded {
		return newPath
	}
	return oldPath
}
