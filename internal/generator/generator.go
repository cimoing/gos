package generator

import (
	"github.com/jake/gola/internal/filesystem"
	"github.com/jake/gola/internal/template"
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
