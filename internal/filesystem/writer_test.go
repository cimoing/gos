package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriterWritesFiles(t *testing.T) {
	root := t.TempDir()
	writer := NewWriter()

	result, err := writer.Write([]File{
		{Path: "cmd/api/main.go", Content: []byte("package main\nfunc main( ){}\n")},
		{Path: "README.md", Content: []byte("# App\n")},
	}, WriteOptions{Root: root})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if len(result.Created) != 2 {
		t.Fatalf("created files = %v, want 2 files", result.Created)
	}

	content, err := os.ReadFile(filepath.Join(root, "cmd", "api", "main.go"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "package main\n\nfunc main() {}\n" {
		t.Fatalf("formatted Go content = %q", string(content))
	}
}

func TestWriterDetectsConflicts(t *testing.T) {
	root := t.TempDir()
	writer := NewWriter()

	_, err := writer.Write([]File{{Path: "README.md", Content: []byte("first")}}, WriteOptions{Root: root})
	if err != nil {
		t.Fatalf("initial Write() error = %v", err)
	}

	_, err = writer.Write([]File{{Path: "README.md", Content: []byte("second")}}, WriteOptions{Root: root})
	if err == nil {
		t.Fatalf("Write() error = nil, want conflict")
	}
}

func TestWriterSkipsExistingOptionalFiles(t *testing.T) {
	root := t.TempDir()
	writer := NewWriter()

	_, err := writer.Write([]File{{Path: "internal/domain/user/entity.go", Content: []byte("first")}}, WriteOptions{Root: root})
	if err != nil {
		t.Fatalf("initial Write() error = %v", err)
	}

	result, err := writer.Write([]File{
		{Path: "internal/domain/user/entity.go", Content: []byte("second"), SkipIfExists: true},
		{Path: "internal/domain/user/repository.go", Content: []byte("repo")},
	}, WriteOptions{Root: root})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if len(result.Skipped) != 1 {
		t.Fatalf("skipped files = %v, want one skipped file", result.Skipped)
	}
	if len(result.Created) != 1 {
		t.Fatalf("created files = %v, want one created file", result.Created)
	}
}

func TestWriterOverwritesAllowedFile(t *testing.T) {
	root := t.TempDir()
	writer := NewWriter()

	_, err := writer.Write([]File{{Path: "router.go", Content: []byte("package http\n")}}, WriteOptions{Root: root})
	if err != nil {
		t.Fatalf("initial Write() error = %v", err)
	}

	result, err := writer.Write([]File{{Path: "router.go", Content: []byte("package api\n"), Overwrite: true}}, WriteOptions{Root: root})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if len(result.Updated) != 1 {
		t.Fatalf("updated files = %v, want one updated file", result.Updated)
	}

	content, err := os.ReadFile(filepath.Join(root, "router.go"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "package api\n" {
		t.Fatalf("content = %q, want overwritten content", string(content))
	}
}

func TestWriterDryRunDoesNotWrite(t *testing.T) {
	root := t.TempDir()
	writer := NewWriter()

	result, err := writer.Write([]File{{Path: "README.md", Content: []byte("# App\n")}}, WriteOptions{
		Root:   root,
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if len(result.Created) != 1 {
		t.Fatalf("created files = %v, want one planned file", result.Created)
	}
	if _, err := os.Stat(filepath.Join(root, "README.md")); !os.IsNotExist(err) {
		t.Fatalf("README.md exists after dry run, stat error = %v", err)
	}
}

func TestWriterRejectsEscapingPaths(t *testing.T) {
	root := t.TempDir()
	writer := NewWriter()

	_, err := writer.Write([]File{{Path: "../outside.txt", Content: []byte("nope")}}, WriteOptions{Root: root})
	if err == nil {
		t.Fatalf("Write() error = nil, want escaping path error")
	}
}
