package command

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jake/gola/internal/generator"
	"github.com/jake/gola/internal/project"
	"github.com/jake/gola/internal/scaffold"
)

func runMakeTest(ctx context.Context, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("make:test", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var opts scaffold.MakeTestOptions
	flags.StringVar(&opts.ModulePath, "module", "", "Go module path for imports")
	flags.BoolVar(&opts.Force, "force", false, "overwrite existing files")
	flags.BoolVar(&opts.DryRun, "dry-run", false, "print planned files without writing them")

	normalizedArgs := moveFlagsBeforePositionals(args)
	if err := flags.Parse(normalizedArgs); err != nil {
		return err
	}

	remaining := flags.Args()
	if len(remaining) != 2 {
		return fmt.Errorf("usage: gos make:test <usecase|handler|repository> <name> [--module=<module-path>] [--force] [--dry-run]")
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if opts.ModulePath == "" {
		opts.ModulePath, err = project.DetectModulePath(wd)
		if err != nil {
			return err
		}
	}

	opts.TargetDir = wd
	opts.Kind = remaining[0]
	opts.Name = filepath.ToSlash(remaining[1])

	gen := scaffold.NewCodeGenerator(generator.Default())
	result, err := gen.GenerateTest(ctx, opts)
	if err != nil {
		return err
	}

	if opts.DryRun {
		fmt.Fprintf(stdout, "dry run: test %s %s\n", opts.Kind, opts.Name)
	} else {
		fmt.Fprintf(stdout, "created test: %s %s\n", opts.Kind, opts.Name)
	}
	printResult(stdout, result)
	return nil
}
