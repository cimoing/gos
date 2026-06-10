package command

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cimoing/gos/internal/generator"
	"github.com/cimoing/gos/internal/scaffold"
)

func runMakeUsecase(ctx context.Context, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("make:usecase", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var opts scaffold.MakeUsecaseOptions
	flags.BoolVar(&opts.Force, "force", false, "overwrite existing files")
	flags.BoolVar(&opts.DryRun, "dry-run", false, "print planned files without writing them")

	normalizedArgs := moveFlagsBeforePositionals(args)
	if err := flags.Parse(normalizedArgs); err != nil {
		return err
	}

	remaining := flags.Args()
	if len(remaining) != 1 {
		return fmt.Errorf("usage: gos make:usecase <module>/<action> [--force] [--dry-run]")
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	opts.TargetDir = wd
	opts.Name = filepath.ToSlash(remaining[0])

	gen := scaffold.NewCodeGenerator(generator.Default())
	result, err := gen.GenerateUsecase(ctx, opts)
	if err != nil {
		return err
	}

	if opts.DryRun {
		fmt.Fprintf(stdout, "dry run: usecase %s\n", opts.Name)
	} else {
		fmt.Fprintf(stdout, "created usecase: %s\n", opts.Name)
	}
	printResult(stdout, result)
	return nil
}
