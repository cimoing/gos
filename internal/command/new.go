package command

import (
	"context"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/jake/gola/internal/generator"
	"github.com/jake/gola/internal/scaffold"
)

func runNew(ctx context.Context, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("new", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var opts scaffold.NewProjectOptions
	flags.StringVar(&opts.ModulePath, "module", "", "Go module path for the generated project")
	flags.StringVar(&opts.Template, "template", "api-clean", "project template")
	flags.BoolVar(&opts.WithOpenTelemetry, "with-otel", false, "include optional OpenTelemetry tracing support")
	flags.BoolVar(&opts.Force, "force", false, "overwrite existing files")
	flags.BoolVar(&opts.DryRun, "dry-run", false, "print planned files without writing them")

	normalizedArgs := moveFlagsBeforePositionals(args)
	if err := flags.Parse(normalizedArgs); err != nil {
		return err
	}

	remaining := flags.Args()
	if len(remaining) != 1 {
		return fmt.Errorf("usage: gos new <project> [--module=<module>] [--template=api-clean|api-minimal] [--with-otel] [--force] [--dry-run]")
	}

	projectPath, err := filepath.Abs(remaining[0])
	if err != nil {
		return err
	}

	if opts.ModulePath == "" {
		opts.ModulePath = inferModulePath(remaining[0])
	}
	opts.ProjectName = filepath.Base(projectPath)
	opts.TargetDir = projectPath

	gen := scaffold.NewProjectGenerator(generator.Default())
	result, err := gen.Generate(ctx, opts)
	if err != nil {
		return err
	}

	if opts.DryRun {
		fmt.Fprintf(stdout, "dry run: %s\n", opts.ProjectName)
	} else {
		fmt.Fprintf(stdout, "created project: %s\n", opts.TargetDir)
	}
	printResult(stdout, result)

	return nil
}

func moveFlagsBeforePositionals(args []string) []string {
	var flags []string
	var positionals []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			if !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				flags = append(flags, args[i+1])
				i++
			}
			continue
		}
		positionals = append(positionals, arg)
	}

	normalized := make([]string, 0, len(args))
	normalized = append(normalized, flags...)
	normalized = append(normalized, positionals...)
	return normalized
}

func inferModulePath(project string) string {
	name := filepath.Base(project)
	name = strings.TrimSpace(name)
	if name == "." || name == string(filepath.Separator) || name == "" {
		return "myapp"
	}
	return name
}
