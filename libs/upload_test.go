package libs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// errReader fails immediately so we can exercise the failure branch of
// SaveUploaded without touching a real upload.
type errReader struct{}

func (errReader) Read(p []byte) (int, error) {
	return 0, errors.New("boom")
}

func countFiles(t *testing.T, dir string) int {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%q) failed: %v", dir, err)
	}
	return len(entries)
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) failed: %v", path, err)
	}
	return string(data)
}

// EnsureDir must create the whole missing path and stay happy if it already
// exists (the upload flow must not require the caller to pre-create the dir).
func TestEnsureDir_CreatesNestedMissingDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "a", "b", "static", "upload")

	if err := EnsureDir(dir); err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("expected dir to exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", dir)
	}

	// Idempotent: a second call on an existing dir must not error.
	if err := EnsureDir(dir); err != nil {
		t.Fatalf("EnsureDir on existing dir failed: %v", err)
	}
}

// Regression: a missing upload directory must be created automatically instead
// of failing the whole upload.
func TestSaveUploaded_CreatesDirWhenMissing(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "static", "upload")
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("precondition failed: dir should not exist yet")
	}

	savedPath, err := SaveUploaded(dir, "cover.png", strings.NewReader("payload"))
	if err != nil {
		t.Fatalf("SaveUploaded failed: %v", err)
	}
	if filepath.Dir(savedPath) != filepath.Clean(dir) {
		t.Fatalf("file saved outside target dir: got %q want dir %q", savedPath, dir)
	}
	if got := readFile(t, savedPath); got != "payload" {
		t.Fatalf("unexpected file content: got %q", got)
	}
}

// Regression: uploading two files that share the same original name must never
// overwrite the first one.
func TestSaveUploaded_SameNameNoOverwrite(t *testing.T) {
	dir := t.TempDir()

	p1, err := SaveUploaded(dir, "pic.jpg", strings.NewReader("first"))
	if err != nil {
		t.Fatalf("first SaveUploaded failed: %v", err)
	}
	p2, err := SaveUploaded(dir, "pic.jpg", strings.NewReader("second"))
	if err != nil {
		t.Fatalf("second SaveUploaded failed: %v", err)
	}

	if p1 == p2 {
		t.Fatalf("expected distinct paths for same original name, both were %q", p1)
	}
	if c := countFiles(t, dir); c != 2 {
		t.Fatalf("expected 2 stored files, found %d", c)
	}
	if got := readFile(t, p1); got != "first" {
		t.Fatalf("first file was overwritten/corrupted: got %q", got)
	}
	if got := readFile(t, p2); got != "second" {
		t.Fatalf("second file content wrong: got %q", got)
	}
	for _, p := range []string{p1, p2} {
		if !strings.HasSuffix(p, ".jpg") {
			t.Fatalf("expected .jpg extension preserved, got %q", p)
		}
	}
}

// UniqueUploadPath must skip names that already exist on disk.
func TestUniqueUploadPath_SkipsExistingFile(t *testing.T) {
	dir := t.TempDir()

	p1, err := UniqueUploadPath(dir, "pic.jpg")
	if err != nil {
		t.Fatalf("UniqueUploadPath failed: %v", err)
	}
	if err := os.WriteFile(p1, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	p2, err := UniqueUploadPath(dir, "pic.jpg")
	if err != nil {
		t.Fatalf("UniqueUploadPath failed: %v", err)
	}
	if p1 == p2 {
		t.Fatalf("expected a new path when the file already exists, got %q twice", p1)
	}
	if _, err := os.Stat(p2); !os.IsNotExist(err) {
		t.Fatalf("returned path should not exist yet: %q", p2)
	}
}

// A malicious or messy original filename must not let the saved file escape the
// upload directory, and the extension must be preserved.
func TestSaveUploaded_SanitizesUnsafeName(t *testing.T) {
	dir := t.TempDir()

	savedPath, err := SaveUploaded(dir, "../../../etc/pas swd.png", strings.NewReader("x"))
	if err != nil {
		t.Fatalf("SaveUploaded failed: %v", err)
	}
	if filepath.Dir(savedPath) != filepath.Clean(dir) {
		t.Fatalf("file escaped upload dir: %q", savedPath)
	}
	name := filepath.Base(savedPath)
	if strings.Contains(name, "..") || strings.ContainsAny(name, "/\\ ") {
		t.Fatalf("unsafe characters survived sanitisation: %q", name)
	}
	if !strings.HasSuffix(savedPath, ".png") {
		t.Fatalf("expected .png extension, got %q", savedPath)
	}
}

func TestSanitizeFileName(t *testing.T) {
	cases := map[string]string{
		"../../evil file.png": "evil_file.png",
		"a/b/c.JPG":           "c.JPG",
		".gitignore":          "gitignore",
		"..":                  "",
		"normal-name_1.gif":   "normal-name_1.gif",
	}
	for in, want := range cases {
		if got := SanitizeFileName(in); got != want {
			t.Errorf("SanitizeFileName(%q) = %q, want %q", in, got, want)
		}
	}
}

// If writing the bytes fails, no partial file may be left behind (so a failed
// upload can never be mistaken for a valid image).
func TestSaveUploaded_RemovesPartialFileOnError(t *testing.T) {
	dir := t.TempDir()

	if _, err := SaveUploaded(dir, "broken.png", errReader{}); err == nil {
		t.Fatalf("expected an error from a failing reader")
	}
	if c := countFiles(t, dir); c != 0 {
		t.Fatalf("expected no leftover files after a failed save, found %d", c)
	}
}

// Update flow: when the user does NOT reselect an image, the old image is kept.
func TestResolveImageURL_KeepsOldWhenNotUploaded(t *testing.T) {
	old := "static/upload/old.jpg"
	if got := ResolveImageURL(old, "static/upload/new.jpg", false); got != old {
		t.Fatalf("expected old image kept, got %q", got)
	}
	// Even if no new path was produced, the old one must be preserved.
	if got := ResolveImageURL(old, "", false); got != old {
		t.Fatalf("expected old image kept, got %q", got)
	}
}

// Update flow: when the user DOES upload a new image, switch to it.
func TestResolveImageURL_UsesNewWhenUploaded(t *testing.T) {
	newPath := "static/upload/new.jpg"
	if got := ResolveImageURL("static/upload/old.jpg", newPath, true); got != newPath {
		t.Fatalf("expected new image used, got %q", got)
	}
}
