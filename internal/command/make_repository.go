package command

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/cimoing/gos/internal/generator"
	"github.com/cimoing/gos/internal/project"
	"github.com/cimoing/gos/internal/scaffold"
)

func runMakeRepository(ctx context.Context, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("make:repository", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var opts scaffold.MakeRepositoryOptions
	flags.StringVar(&opts.ModulePath, "module", "", "Go module path for imports")
	flags.StringVar(&opts.DB, "db", "mysql", "database implementation")
	flags.StringVar(&opts.TableName, "table", "", "database table name")
	flags.StringVar(&opts.Fields, "fields", "", "comma-separated fields, for example name:string,email:string:unique,size=320,json=email_address,default=test,sql=TEXT")
	flags.BoolVar(&opts.WithMigration, "with-migration", false, "generate a create table migration")
	flags.StringVar(&opts.MigrationDir, "migration-dir", "migrations", "migration directory")
	flags.BoolVar(&opts.Register, "register", false, "register repository in internal/app/assembly.go")
	flags.BoolVar(&opts.Force, "force", false, "overwrite existing files")
	flags.BoolVar(&opts.DryRun, "dry-run", false, "print planned files without writing them")

	normalizedArgs := moveFlagsBeforePositionals(args)
	if err := flags.Parse(normalizedArgs); err != nil {
		return err
	}

	remaining := flags.Args()
	if len(remaining) != 1 {
		return fmt.Errorf("usage: gos make:repository <module> [--module=<module-path>] [--db=mysql] [--table=<table>] [--fields=name:string[:unique,nullable,size=320,default=value,sql=TEXT,json=name]] [--with-migration] [--migration-dir=migrations] [--register] [--force] [--dry-run]")
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
	result, err := gen.GenerateRepository(ctx, opts)
	if err != nil {
		return err
	}

	if opts.DryRun {
		fmt.Fprintf(stdout, "dry run: repository %s\n", opts.Name)
	} else {
		fmt.Fprintf(stdout, "created repository: %s\n", opts.Name)
	}
	printResult(stdout, result)
	return nil
}
