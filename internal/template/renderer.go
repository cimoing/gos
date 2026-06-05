package template

import (
	"bytes"
	"io/fs"
	"path/filepath"
	"strings"
	texttemplate "text/template"
)

type Renderer struct{}

func NewRenderer() *Renderer {
	return &Renderer{}
}

func (r *Renderer) RenderString(name string, body string, data any) ([]byte, error) {
	tmpl, err := texttemplate.New(name).Option("missingkey=error").Parse(body)
	if err != nil {
		return nil, err
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, data); err != nil {
		return nil, err
	}

	return rendered.Bytes(), nil
}

func (r *Renderer) RenderFiles(source fs.FS, root string, data any) (map[string][]byte, error) {
	files := make(map[string][]byte)
	err := fs.WalkDir(source, root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		raw, err := fs.ReadFile(source, path)
		if err != nil {
			return err
		}

		outputPath := strings.TrimPrefix(filepath.ToSlash(path), filepath.ToSlash(root)+"/")
		outputPath = strings.TrimSuffix(outputPath, ".tmpl")

		tmpl, err := texttemplate.New(filepath.Base(path)).Option("missingkey=error").Parse(string(raw))
		if err != nil {
			return err
		}

		var rendered bytes.Buffer
		if err := tmpl.Execute(&rendered, data); err != nil {
			return err
		}
		if len(bytes.TrimSpace(rendered.Bytes())) == 0 {
			return nil
		}

		files[outputPath] = rendered.Bytes()
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}
