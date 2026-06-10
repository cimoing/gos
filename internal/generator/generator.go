package generator

import (
	"github.com/cimoing/gos/internal/filesystem"
	"github.com/cimoing/gos/internal/template"
)

type Engine struct {
	Templates *template.Renderer
	Writer    *filesystem.Writer
}

func Default() *Engine {
	return &Engine{
		Templates: template.NewRenderer(),
		Writer:    filesystem.NewWriter(),
	}
}
