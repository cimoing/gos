package command

import (
	"context"
	"fmt"
	"io"
)

const usage = `gos is a Go backend project scaffold.

Usage:
  gos new <project> [--module=<module>] [--template=api-clean|api-minimal] [--with-otel] [--force] [--dry-run]
  gos make:usecase <module>/<action> [--force] [--dry-run]
  gos make:handler <module> [--module=<module-path>] [--register] [--openapi] [--force] [--dry-run]
  gos make:model <module> [--fields=name:string[:unique,nullable,size=320,default=value,sql=TEXT,json=name]] [--force] [--dry-run]
  gos make:repository <module> [--module=<module-path>] [--db=mysql] [--table=<table>] [--fields=name:string[:unique,nullable,size=320,default=value,sql=TEXT,json=name]] [--with-migration] [--migration-dir=migrations] [--register] [--force] [--dry-run]
  gos make:migration <name> [--dir=migrations] [--force] [--dry-run]
  gos make:test <usecase|handler|repository> <name> [--module=<module-path>] [--force] [--dry-run]
  gos make:command <name> [--module=<module-path>] [--register] [--force] [--dry-run]
  gos help
`

func Execute(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		_, _ = fmt.Fprint(stdout, usage)
		return nil
	}

	switch args[0] {
	case "new":
		return runNew(ctx, args[1:], stdout)
	case "make:usecase":
		return runMakeUsecase(ctx, args[1:], stdout)
	case "make:handler":
		return runMakeHandler(ctx, args[1:], stdout)
	case "make:model":
		return runMakeModel(ctx, args[1:], stdout)
	case "make:repository":
		return runMakeRepository(ctx, args[1:], stdout)
	case "make:migration":
		return runMakeMigration(ctx, args[1:], stdout)
	case "make:test":
		return runMakeTest(ctx, args[1:], stdout)
	case "make:command":
		return runMakeCommand(ctx, args[1:], stdout)
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}
}
