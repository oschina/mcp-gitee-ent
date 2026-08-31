package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolveReadableFile(t *testing.T) {
	base := t.TempDir()
	allowedRoot := filepath.Join(base, "allowed")
	outsideRoot := filepath.Join(base, "outside")
	if err := os.MkdirAll(allowedRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outsideRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(AllowedFileRootsEnv, allowedRoot)

	allowedFile := filepath.Join(allowedRoot, "page.md")
	if err := os.WriteFile(allowedFile, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
	resolved, err := ResolveReadableFile(allowedFile)
	if err != nil {
		t.Fatalf("ResolveReadableFile() error = %v", err)
	}
	wantResolved, err := filepath.EvalSymlinks(allowedFile)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != wantResolved {
		t.Fatalf("ResolveReadableFile() = %q, want %q", resolved, wantResolved)
	}

	outsideFile := filepath.Join(outsideRoot, "outside.md")
	if err := os.WriteFile(outsideFile, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveReadableFile(outsideFile); !IsParamError(err) {
		t.Fatalf("ResolveReadableFile(outside) error = %v, want parameter error", err)
	}

	if _, err := ResolveReadableFile("relative.md"); !IsParamError(err) {
		t.Fatalf("ResolveReadableFile(relative) error = %v, want parameter error", err)
	}
}

func TestResolveReadableFileRejectsSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation may require elevated privileges on Windows")
	}

	base := t.TempDir()
	allowedRoot := filepath.Join(base, "allowed")
	outsideRoot := filepath.Join(base, "outside")
	if err := os.MkdirAll(allowedRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outsideRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(AllowedFileRootsEnv, allowedRoot)

	outsideFile := filepath.Join(outsideRoot, "secret.md")
	if err := os.WriteFile(outsideFile, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(allowedRoot, "link.md")
	if err := os.Symlink(outsideFile, link); err != nil {
		t.Fatal(err)
	}

	if _, err := ResolveReadableFile(link); !IsParamError(err) {
		t.Fatalf("ResolveReadableFile(symlink escape) error = %v, want parameter error", err)
	}
}

func TestResolveWritableDirectory(t *testing.T) {
	allowedRoot := t.TempDir()
	t.Setenv(AllowedFileRootsEnv, allowedRoot)

	outputDirectory := filepath.Join(allowedRoot, "exports")
	if err := os.Mkdir(outputDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	resolved, err := ResolveWritableDirectory(outputDirectory)
	if err != nil {
		t.Fatalf("ResolveWritableDirectory() error = %v", err)
	}
	wantResolved, err := filepath.EvalSymlinks(outputDirectory)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != wantResolved {
		t.Fatalf("ResolveWritableDirectory() = %q, want %q", resolved, wantResolved)
	}

	if _, err := ResolveWritableDirectory("exports"); !IsParamError(err) {
		t.Fatalf("ResolveWritableDirectory(relative) error = %v, want parameter error", err)
	}
}
