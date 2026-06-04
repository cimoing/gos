package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectModulePath(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "internal", "command")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/demo\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	modulePath, err := DetectModulePath(nested)
	if err != nil {
		t.Fatalf("DetectModulePath() error = %v", err)
	}
	if modulePath != "example.com/demo" {
		t.Fatalf("module path = %q, want example.com/demo", modulePath)
	}
}

func TestDetectModulePathNotFound(t *testing.T) {
	_, err := DetectModulePath(t.TempDir())
	if err == nil {
		t.Fatalf("DetectModulePath() error = nil, want not found error")
	}
}
