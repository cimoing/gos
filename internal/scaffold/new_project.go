package scaffold

import (
	"context"
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jake/gola/internal/filesystem"
	"github.com/jake/gola/internal/generator"
	"github.com/jake/gola/internal/naming"
)

//go:embed templates/* templates/api-clean/.env.example.tmpl templates/api-clean/.gitignore.tmpl templates/api-clean/.github/workflows/ci.yml.tmpl
var templateFS embed.FS

type NewProjectOptions struct {
	ProjectName string
	ModulePath  string
	Template    string
	TargetDir   string
	Force       bool
	DryRun      bool
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
	if opts.Template == "" {
		opts.Template = "api-clean"
	}
	if opts.Template != "api-clean" {
		return nil, fmt.Errorf("unknown template %q", opts.Template)
	}

	data := projectTemplateData{
		ProjectName: opts.ProjectName,
		ModulePath:  opts.ModulePath,
		AppName:     naming.ToKebab(opts.ProjectName),
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

type projectTemplateData struct {
	ProjectName string
	ModulePath  string
	AppName     string
}
