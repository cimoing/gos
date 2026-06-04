package command

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/jake/gola/internal/generator"
	"github.com/jake/gola/internal/scaffold"
)

func runMakeModel(ctx context.Context, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("make:model", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var opts scaffold.MakeModelOptions
	flags.StringVar(&opts.Fields, "fields", "", "comma-separated fields, for example name:string,email:string:unique,size=320,json=email_address,default=test,sql=TEXT")
	flags.BoolVar(&opts.Force, "force", false, "overwrite existing files")
	flags.BoolVar(&opts.DryRun, "dry-run", false, "print planned files without writing them")

	normalizedArgs := moveFlagsBeforePositionals(args)
	if err := flags.Parse(normalizedArgs); err != nil {
		return err
	}

	remaining := flags.Args()
	if len(remaining) != 1 {
		return fmt.Errorf("usage: gos make:model <module> [--fields=name:string[:unique,nullable,size=320,default=value,sql=TEXT,json=name]] [--force] [--dry-run]")
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	opts.TargetDir = wd
	opts.Name = remaining[0]

	gen := scaffold.NewCodeGenerator(generator.Default())
	result, err := gen.GenerateModel(ctx, opts)
	if err != nil {
		return err
	}

	if opts.DryRun {
		fmt.Fprintf(stdout, "dry run: model %s\n", opts.Name)
	} else {
		fmt.Fprintf(stdout, "created model: %s\n", opts.Name)
	}
	printResult(stdout, result)
	return nil
}
