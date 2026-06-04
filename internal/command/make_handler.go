package command

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jake/gola/internal/generator"
	"github.com/jake/gola/internal/naming"
	"github.com/jake/gola/internal/project"
	"github.com/jake/gola/internal/scaffold"
)

func runMakeHandler(ctx context.Context, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("make:handler", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var opts scaffold.MakeHandlerOptions
	flags.StringVar(&opts.ModulePath, "module", "", "Go module path for imports")
	flags.BoolVar(&opts.Register, "register", false, "register handler routes in internal/interfaces/http/router.go")
	flags.BoolVar(&opts.OpenAPI, "openapi", false, "append handler path to api/openapi.yaml")
	flags.BoolVar(&opts.Force, "force", false, "overwrite existing files")
	flags.BoolVar(&opts.DryRun, "dry-run", false, "print planned files without writing them")

	normalizedArgs := moveFlagsBeforePositionals(args)
	if err := flags.Parse(normalizedArgs); err != nil {
		return err
	}

	remaining := flags.Args()
	if len(remaining) != 1 {
		return fmt.Errorf("usage: gos make:handler <module> [--module=<module-path>] [--register] [--openapi] [--force] [--dry-run]")
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
	result, err := gen.GenerateHandler(ctx, opts)
	if err != nil {
		return err
	}

	if opts.DryRun {
		fmt.Fprintf(stdout, "dry run: handler %s\n", opts.Name)
	} else {
		fmt.Fprintf(stdout, "created handler: %s\n", opts.Name)
	}
	printResult(stdout, result)
	if opts.Register && containsPath(result.Skipped, "internal/interfaces/http/router.go") {
		typeName := naming.ToPascal(opts.Name)
		variableName := naming.ToCamel(opts.Name) + "Handler"
		fmt.Fprintln(stdout, "  hint register routes manually in internal/interfaces/http/router.go:")
		fmt.Fprintf(stdout, "    %s := handler.New%sHandler()\n", variableName, typeName)
		fmt.Fprintf(stdout, "    %s.RegisterRoutes(mux)\n", variableName)
	}
	if opts.OpenAPI && containsPath(result.Skipped, "api/openapi.yaml") {
		routePath := "/" + naming.ToKebab(opts.Name) + "s"
		fmt.Fprintln(stdout, "  hint register OpenAPI path manually in api/openapi.yaml:")
		fmt.Fprintf(stdout, "    %s:\n", routePath)
	}
	return nil
}

func containsPath(paths []string, want string) bool {
	want = filepath.ToSlash(want)
	for _, path := range paths {
		if filepath.ToSlash(path) == want {
			return true
		}
	}
	return false
}
