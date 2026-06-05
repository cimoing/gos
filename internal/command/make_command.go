package command

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/jake/gola/internal/generator"
	"github.com/jake/gola/internal/naming"
	"github.com/jake/gola/internal/project"
	"github.com/jake/gola/internal/scaffold"
)

func runMakeCommand(ctx context.Context, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("make:command", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var opts scaffold.MakeCommandOptions
	flags.StringVar(&opts.ModulePath, "module", "", "Go module path for imports")
	flags.BoolVar(&opts.Register, "register", false, "register command in cmd/api/main.go")
	flags.BoolVar(&opts.Force, "force", false, "overwrite existing files")
	flags.BoolVar(&opts.DryRun, "dry-run", false, "print planned files without writing them")

	normalizedArgs := moveFlagsBeforePositionals(args)
	if err := flags.Parse(normalizedArgs); err != nil {
		return err
	}

	remaining := flags.Args()
	if len(remaining) != 1 {
		return fmt.Errorf("usage: gos make:command <name> [--module=<module-path>] [--register] [--force] [--dry-run]")
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
	opts.Name = remaining[0]

	gen := scaffold.NewCodeGenerator(generator.Default())
	result, err := gen.GenerateCommand(ctx, opts)
	if err != nil {
		return err
	}

	if opts.DryRun {
		fmt.Fprintf(stdout, "dry run: command %s\n", naming.ToKebab(opts.Name))
	} else {
		fmt.Fprintf(stdout, "created command: %s\n", naming.ToKebab(opts.Name))
	}
	printResult(stdout, result)
	if opts.Register && containsPath(result.Skipped, "cmd/api/main.go") {
		fmt.Fprintln(stdout, "  hint register command manually in cmd/api/main.go:")
		fmt.Fprintf(stdout, "    rootCmd.AddCommand(appcommand.New%sCommand())\n", naming.ToPascal(opts.Name))
	}
	return nil
}
