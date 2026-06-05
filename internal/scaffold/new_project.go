package scaffold

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jake/gola/internal/filesystem"
	"github.com/jake/gola/internal/generator"
	"github.com/jake/gola/internal/naming"
)

const DefaultProjectTemplate = "api-clean"

//go:embed templates/* templates/api-clean/.env.example.tmpl templates/api-clean/.gitignore.tmpl templates/api-clean/.github/workflows/ci.yml.tmpl templates/api-minimal/.gitignore.tmpl
var templateFS embed.FS

type NewProjectOptions struct {
	ProjectName       string
	ModulePath        string
	Template          string
	TargetDir         string
	WithOpenTelemetry bool
	Force             bool
	DryRun            bool
}

type ProjectGenerator struct {
	engine *generator.Engine
}

func NewProjectGenerator(engine *generator.Engine) *ProjectGenerator {
	return &ProjectGenerator{engine: engine}
}

func (g *ProjectGenerator) Generate(ctx context.Context, opts NewProjectOptions) (*filesystem.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(opts.ProjectName) == "" {
		return nil, fmt.Errorf("project name is required")
	}
	if strings.TrimSpace(opts.ModulePath) == "" {
		return nil, fmt.Errorf("module path is required")
	}
	opts.Template = strings.TrimSpace(opts.Template)
	if opts.Template == "" {
		opts.Template = DefaultProjectTemplate
	}
	if !projectTemplateExists(opts.Template) {
		return nil, fmt.Errorf("unknown template %q (available: %s)", opts.Template, strings.Join(SupportedProjectTemplates(), ", "))
	}

	data := projectTemplateData{
		ProjectName:       opts.ProjectName,
		ModulePath:        opts.ModulePath,
		AppName:           naming.ToKebab(opts.ProjectName),
		WithOpenTelemetry: opts.WithOpenTelemetry,
	}

	rendered, err := g.engine.Templates.RenderFiles(templateFS, filepath.ToSlash(filepath.Join("templates", opts.Template)), data)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(rendered))
	for path := range rendered {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	files := make([]filesystem.File, 0, len(paths))
	for _, path := range paths {
		files = append(files, filesystem.File{
			Path:    path,
			Content: rendered[path],
		})
	}

	return g.engine.Writer.Write(files, filesystem.WriteOptions{
		Root:   opts.TargetDir,
		Force:  opts.Force,
		DryRun: opts.DryRun,
	})
}

func SupportedProjectTemplates() []string {
	entries, err := fs.ReadDir(templateFS, "templates")
	if err != nil {
		return []string{DefaultProjectTemplate}
	}

	templates := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			templates = append(templates, entry.Name())
		}
	}
	sort.Strings(templates)
	return templates
}

func projectTemplateExists(name string) bool {
	for _, template := range SupportedProjectTemplates() {
		if template == name {
			return true
		}
	}
	return false
}

type projectTemplateData struct {
	ProjectName       string
	ModulePath        string
	AppName           string
	WithOpenTelemetry bool
}
