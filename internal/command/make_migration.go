package command

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jake/gola/internal/generator"
	"github.com/jake/gola/internal/scaffold"
)

func runMakeMigration(ctx context.Context, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("make:migration", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var opts scaffold.MakeMigrationOptions
	flags.StringVar(&opts.Dir, "dir", "migrations", "migration directory")
	flags.BoolVar(&opts.Force, "force", false, "overwrite existing files")
	flags.BoolVar(&opts.DryRun, "dry-run", false, "print planned files without writing them")

	normalizedArgs := moveFlagsBeforePositionals(args)
	if err := flags.Parse(normalizedArgs); err != nil {
		return err
	}

	remaining := flags.Args()
	if len(remaining) != 1 {
		return fmt.Errorf("usage: gos make:migration <name> [--dir=migrations] [--force] [--dry-run]")
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	opts.TargetDir = wd
	opts.Name = filepath.ToSlash(remaining[0])

	gen := scaffold.NewCodeGenerator(generator.Default())
	result, err := gen.GenerateMigration(ctx, opts)
	if err != nil {
		return err
	}

	if opts.DryRun {
		fmt.Fprintf(stdout, "dry run: migration %s\n", opts.Name)
	} else {
		fmt.Fprintf(stdout, "created migration: %s\n", opts.Name)
	}
	printResult(stdout, result)
	return nil
}
