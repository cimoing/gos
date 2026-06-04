package filesystem

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	Path         string
	Content      []byte
	SkipIfExists bool
	Overwrite    bool
}

type WriteOptions struct {
	Root   string
	Force  bool
	DryRun bool
}

type Result struct {
	Created []string
	Updated []string
	Skipped []string
}

type Writer struct{}

func NewWriter() *Writer {
	return &Writer{}
}

func (w *Writer) Write(files []File, opts WriteOptions) (*Result, error) {
	root, err := filepath.Abs(opts.Root)
	if err != nil {
		return nil, err
	}

	result := &Result{}
	for _, file := range files {
		target, err := safeJoin(root, file.Path)
		if err != nil {
			return nil, err
		}

		content := formatIfGo(target, file.Content)
		rel, err := filepath.Rel(root, target)
		if err != nil {
			return nil, err
		}
		rel = filepath.ToSlash(rel)

		exists, err := fileExists(target)
		if err != nil {
			return nil, err
		}
		if exists && !opts.Force {
			if file.SkipIfExists {
				result.Skipped = append(result.Skipped, rel)
				continue
			}
			if !file.Overwrite {
				return nil, fmt.Errorf("file already exists: %s", rel)
			}
		}

		if opts.DryRun {
			if exists {
				result.Updated = append(result.Updated, rel)
			} else {
				result.Created = append(result.Created, rel)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(target, content, 0o644); err != nil {
			return nil, err
		}

		if exists {
			result.Updated = append(result.Updated, rel)
		} else {
			result.Created = append(result.Created, rel)
		}
	}

	return result, nil
}

func safeJoin(root, name string) (string, error) {
	if filepath.IsAbs(name) {
		return "", fmt.Errorf("template path must be relative: %s", name)
	}

	target, err := filepath.Abs(filepath.Join(root, name))
	if err != nil {
		return "", err
	}

	cleanRoot := filepath.Clean(root)
	cleanTarget := filepath.Clean(target)
	if cleanTarget != cleanRoot && !strings.HasPrefix(cleanTarget, cleanRoot+string(filepath.Separator)) {
		return "", fmt.Errorf("template path escapes target directory: %s", name)
	}

	return target, nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func formatIfGo(path string, content []byte) []byte {
	if filepath.Ext(path) != ".go" {
		return content
	}

	formatted, err := format.Source(content)
	if err != nil {
		return bytes.TrimSpace(content)
	}
	return formatted
}
