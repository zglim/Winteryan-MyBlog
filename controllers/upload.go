package controllers

import (
	"crypto/rand"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/astaxie/beego"
)

// UploadDir is the default directory for uploaded files.
var UploadDir = "static/upload"

// ensureUploadDir creates the upload directory (and parents) if it does not exist.
func ensureUploadDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// safeFileName generates a unique filename to avoid overwriting existing files.
// Format: <timestamp>_<random8>.<ext>
func safeFileName(original string) string {
	ext := filepath.Ext(original)
	// Sanitise extension to prevent path traversal
	ext = strings.ToLower(ext)
	if strings.ContainsAny(ext, "/\\") {
		ext = ""
	}

	ts := time.Now().Format("20060102150405.000")
	randBytes := make([]byte, 4)
	_, _ = rand.Read(randBytes)
	randHex := fmt.Sprintf("%x", randBytes)

	return fmt.Sprintf("%s_%s%s", ts, randHex, ext)
}

// saveUploadedFile saves a multipart file to the given destination path.
func saveUploadedFile(file multipart.File, dst string) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	return err
}

// uploadResult holds the outcome of a successful upload.
type uploadResult struct {
	// RelPath is the relative path stored in the database, e.g. "static/upload/xxx.jpg"
	RelPath string
	// AbsPath is the absolute filesystem path of the saved file.
	AbsPath string
}

// handleUpload performs the full upload pipeline:
//  1. Ensure upload directory exists.
//  2. Generate a safe, unique filename.
//  3. Save the file to disk.
//
// It returns the upload result or an error.
func handleUpload(file multipart.File, header *multipart.FileHeader) (*uploadResult, error) {
	if err := ensureUploadDir(UploadDir); err != nil {
		return nil, fmt.Errorf("创建上传目录失败: %v", err)
	}

	newName := safeFileName(header.Filename)
	relPath := filepath.ToSlash(filepath.Join(UploadDir, newName))
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		return nil, fmt.Errorf("解析路径失败: %v", err)
	}

	if err := saveUploadedFile(file, absPath); err != nil {
		return nil, fmt.Errorf("保存文件失败: %v", err)
	}

	beego.Info("文件上传成功: ", relPath)
	return &uploadResult{RelPath: relPath, AbsPath: absPath}, nil
}
