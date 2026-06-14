package controllers

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------- ensureUploadDir ----------

func TestEnsureUploadDir_CreatesMissingDir(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "upload_test_ensure_dir", "sub")
	defer os.RemoveAll(filepath.Join(os.TempDir(), "upload_test_ensure_dir"))

	if err := ensureUploadDir(dir); err != nil {
		t.Fatalf("ensureUploadDir returned error: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("path exists but is not a directory")
	}
}

func TestEnsureUploadDir_ExistingDirNoError(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "upload_test_existing_dir")
	os.MkdirAll(dir, 0755)
	defer os.RemoveAll(dir)

	if err := ensureUploadDir(dir); err != nil {
		t.Fatalf("ensureUploadDir on existing dir returned error: %v", err)
	}
}

// ---------- safeFileName ----------

func TestSafeFileName_PreservesExtension(t *testing.T) {
	name := safeFileName("photo.jpg")
	if !strings.HasSuffix(name, ".jpg") {
		t.Fatalf("expected .jpg extension, got %s", name)
	}
}

func TestSafeFileName_DifferentForSameInput(t *testing.T) {
	n1 := safeFileName("same.png")
	// Small sleep isn't needed because random bytes differ each call.
	n2 := safeFileName("same.png")
	if n1 == n2 {
		t.Fatalf("safeFileName should generate different names for same input, got %s twice", n1)
	}
}

func TestSafeFileName_NoPathTraversal(t *testing.T) {
	name := safeFileName("../../etc/passwd")
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
		t.Fatalf("safeFileName should strip path components, got %s", name)
	}
}

func TestSafeFileName_EmptyExtension(t *testing.T) {
	name := safeFileName("noext")
	if strings.Contains(name, ".") {
		// Should not have an extension dot except if ext was empty.
		// Actually our format is ts_randomhex + ext, ext is "" so no dot.
		if strings.HasSuffix(name, ".") {
			t.Fatalf("unexpected trailing dot in %s", name)
		}
	}
}

// ---------- saveUploadedFile ----------

func TestSaveUploadedFile(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "upload_test_save")
	os.MkdirAll(dir, 0755)
	defer os.RemoveAll(dir)

	content := []byte("hello world image data")
	file := newFakeMultipartFile(content)

	dst := filepath.Join(dir, "test.jpg")
	if err := saveUploadedFile(file, dst); err != nil {
		t.Fatalf("saveUploadedFile failed: %v", err)
	}

	saved, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("could not read saved file: %v", err)
	}
	if !bytes.Equal(saved, content) {
		t.Fatalf("saved content mismatch: got %q, want %q", saved, content)
	}
}

func TestSaveUploadedFile_InvalidPath(t *testing.T) {
	file := newFakeMultipartFile([]byte("data"))
	err := saveUploadedFile(file, "/nonexistent_dir_xyz/sub/test.jpg")
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}

// ---------- handleUpload (full pipeline) ----------

func TestHandleUpload_Success(t *testing.T) {
	// Use a temp upload dir.
	dir := filepath.Join(os.TempDir(), "upload_test_handle")
	defer os.RemoveAll(dir)

	oldDir := UploadDir
	UploadDir = dir
	defer func() { UploadDir = oldDir }()

	content := []byte("fake image bytes")
	file := newFakeMultipartFile(content)
	header := &multipart.FileHeader{Filename: "test.jpg"}

	result, err := handleUpload(file, header)
	if err != nil {
		t.Fatalf("handleUpload failed: %v", err)
	}

	// Check rel path starts with upload dir
	if !strings.HasPrefix(result.RelPath, filepath.ToSlash(dir)) {
		t.Fatalf("RelPath should start with upload dir, got %s", result.RelPath)
	}
	if !strings.HasSuffix(result.RelPath, ".jpg") {
		t.Fatalf("RelPath should have .jpg extension, got %s", result.RelPath)
	}

	// Verify file exists on disk with correct content
	saved, err := os.ReadFile(result.AbsPath)
	if err != nil {
		t.Fatalf("could not read saved file: %v", err)
	}
	if !bytes.Equal(saved, content) {
		t.Fatalf("saved content mismatch")
	}
}

func TestHandleUpload_CreatesDirIfMissing(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "upload_test_missing_dir", "nested")
	defer os.RemoveAll(filepath.Join(os.TempDir(), "upload_test_missing_dir"))

	oldDir := UploadDir
	UploadDir = dir
	defer func() { UploadDir = oldDir }()

	file := newFakeMultipartFile([]byte("data"))
	header := &multipart.FileHeader{Filename: "img.png"}

	_, err := handleUpload(file, header)
	if err != nil {
		t.Fatalf("handleUpload should create missing dir, got error: %v", err)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Fatal("upload dir was not created")
	}
}

// TestHandleUpload_SameFilenameNoOverwrite verifies that uploading two files
// with the same original name does not cause the second to overwrite the first.
func TestHandleUpload_SameFilenameNoOverwrite(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "upload_test_nooverwrite")
	defer os.RemoveAll(dir)

	oldDir := UploadDir
	UploadDir = dir
	defer func() { UploadDir = oldDir }()

	content1 := []byte("first image content")
	content2 := []byte("second image content")

	file1 := newFakeMultipartFile(content1)
	header1 := &multipart.FileHeader{Filename: "duplicate.jpg"}

	result1, err := handleUpload(file1, header1)
	if err != nil {
		t.Fatalf("first upload failed: %v", err)
	}

	file2 := newFakeMultipartFile(content2)
	header2 := &multipart.FileHeader{Filename: "duplicate.jpg"}

	result2, err := handleUpload(file2, header2)
	if err != nil {
		t.Fatalf("second upload failed: %v", err)
	}

	// Paths must differ
	if result1.RelPath == result2.RelPath {
		t.Fatal("same original filename should produce different saved paths")
	}

	// Both files must still exist with their original content
	saved1, _ := os.ReadFile(result1.AbsPath)
	saved2, _ := os.ReadFile(result2.AbsPath)
	if !bytes.Equal(saved1, content1) {
		t.Fatal("first file was overwritten")
	}
	if !bytes.Equal(saved2, content2) {
		t.Fatal("second file content mismatch")
	}
}

// ---------- Keep-old-image / replace-new-image logic ----------

// simulateUpdateImgurl mirrors the controller logic for "keep old or replace with new".
func simulateUpdateImgurl(oldImgurl string, hasNewFile bool, newFile multipart.File, newHeader *multipart.FileHeader) (string, error) {
	imgurl := oldImgurl
	if hasNewFile {
		result, err := handleUpload(newFile, newHeader)
		if err != nil {
			return oldImgurl, err
		}
		imgurl = result.RelPath
	}
	return imgurl, nil
}

func TestKeepOldImage_WhenNoNewUpload(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "upload_test_keepold")
	defer os.RemoveAll(dir)

	oldDir := UploadDir
	UploadDir = dir
	defer func() { UploadDir = oldDir }()

	oldImgurl := "static/upload/old_image.jpg"
	result, err := simulateUpdateImgurl(oldImgurl, false, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != oldImgurl {
		t.Fatalf("expected old imgurl %s, got %s", oldImgurl, result)
	}
}

func TestReplaceNewImage_WhenNewUpload(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "upload_test_replace")
	defer os.RemoveAll(dir)

	oldDir := UploadDir
	UploadDir = dir
	defer func() { UploadDir = oldDir }()

	oldImgurl := "static/upload/old_image.jpg"
	file := newFakeMultipartFile([]byte("new image"))
	header := &multipart.FileHeader{Filename: "new.jpg"}

	result, err := simulateUpdateImgurl(oldImgurl, true, file, header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == oldImgurl {
		t.Fatal("expected new imgurl, got old one")
	}
	if !strings.HasSuffix(result, ".jpg") {
		t.Fatalf("expected .jpg extension, got %s", result)
	}
}

func TestReplaceNewImage_UploadFailsKeepsOld(t *testing.T) {
	// Point UploadDir to a path that will fail to create (e.g., under a file).
	blocker := filepath.Join(os.TempDir(), "upload_test_blocker")
	os.WriteFile(blocker, []byte("block"), 0644)
	defer os.Remove(blocker)

	oldDir := UploadDir
	UploadDir = filepath.Join(blocker, "impossible")
	defer func() { UploadDir = oldDir }()

	oldImgurl := "static/upload/old_image.jpg"
	file := newFakeMultipartFile([]byte("new image"))
	header := &multipart.FileHeader{Filename: "new.jpg"}

	result, err := simulateUpdateImgurl(oldImgurl, true, file, header)
	if err == nil {
		t.Fatal("expected error when upload dir cannot be created")
	}
	if result != oldImgurl {
		t.Fatalf("on upload failure, should keep old imgurl %s, got %s", oldImgurl, result)
	}
}

// ---------- helpers ----------

// fakeMultipartFile wraps a byte slice as a multipart.File for testing.
type fakeMultipartFile struct {
	*bytes.Reader
}

func newFakeMultipartFile(data []byte) *fakeMultipartFile {
	return &fakeMultipartFile{Reader: bytes.NewReader(data)}
}

func (f *fakeMultipartFile) Close() error { return nil }

func (f *fakeMultipartFile) ReadAt(p []byte, off int64) (n int, err error) {
	return f.Reader.ReadAt(p, off)
}

func (f *fakeMultipartFile) Seek(offset int64, whence int) (int64, error) {
	return f.Reader.Seek(offset, whence)
}

// Verify interface compliance
var _ multipart.File = (*fakeMultipartFile)(nil)
var _ io.ReaderAt = (*fakeMultipartFile)(nil)
var _ io.Seeker = (*fakeMultipartFile)(nil)
