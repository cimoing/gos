package command

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
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
  gos version
  gos completion <bash|zsh|fish|powershell>
  gos help
`

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func Execute(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	cmd := NewRootCommand(ctx, stdout, stderr)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func NewRootCommand(ctx context.Context, stdout, stderr io.Writer) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "gos",
		Short:         "Go backend project scaffold",
		Long:          usage,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _ = fmt.Fprint(stdout, usage)
			return nil
		},
	}
	rootCmd.SetContext(ctx)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)

	rootCmd.AddCommand(rawCommand("new", "Create a new project", usageNew, runNew))
	rootCmd.AddCommand(rawCommand("make:usecase", "Generate a usecase", usageMakeUsecase, runMakeUsecase))
	rootCmd.AddCommand(rawCommand("make:handler", "Generate an HTTP handler", usageMakeHandler, runMakeHandler))
	rootCmd.AddCommand(rawCommand("make:model", "Generate a domain model", usageMakeModel, runMakeModel))
	rootCmd.AddCommand(rawCommand("make:repository", "Generate a repository", usageMakeRepository, runMakeRepository))
	rootCmd.AddCommand(rawCommand("make:migration", "Generate migration files", usageMakeMigration, runMakeMigration))
	rootCmd.AddCommand(rawCommand("make:test", "Generate test scaffolding", usageMakeTest, runMakeTest))
	rootCmd.AddCommand(rawCommand("make:command", "Generate a Cobra command script", usageMakeCommand, runMakeCommand))
	rootCmd.AddCommand(newVersionCommand(stdout))
	rootCmd.AddCommand(newCompletionCommand(stdout))

	return rootCmd
}

func newVersionCommand(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:           "version",
		Short:         "Print gos version information",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(stdout, "gos %s\ncommit %s\nbuilt %s\n", Version, Commit, BuildDate)
			return nil
		},
	}
}

func newCompletionCommand(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:           "completion <bash|zsh|fish|powershell>",
		Short:         "Generate shell completion script",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletionV2(stdout, true)
			case "zsh":
				return cmd.Root().GenZshCompletion(stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletion(stdout)
			default:
				return fmt.Errorf("unsupported shell %q (supported: bash, zsh, fish, powershell)", args[0])
			}
		},
	}
}

type commandRunner func(context.Context, []string, io.Writer) error

func rawCommand(use string, short string, usageText string, run commandRunner) *cobra.Command {
	return &cobra.Command{
		Use:                use,
		Short:              short,
		Long:               usageText,
		DisableFlagParsing: true,
		SilenceUsage:       true,
		SilenceErrors:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if wantsHelp(args) {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), usageText)
				return nil
			}
			return run(cmd.Context(), args, cmd.OutOrStdout())
		},
	}
}

func wantsHelp(args []string) bool {
	return len(args) == 1 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h")
}

const usageNew = `Usage:
  gos new <project> [--module=<module>] [--template=api-clean|api-minimal] [--with-otel] [--force] [--dry-run]`

const usageMakeUsecase = `Usage:
  gos make:usecase <module>/<action> [--force] [--dry-run]`

const usageMakeHandler = `Usage:
  gos make:handler <module> [--module=<module-path>] [--register] [--openapi] [--force] [--dry-run]`

const usageMakeModel = `Usage:
  gos make:model <module> [--fields=name:string[:unique,nullable,size=320,default=value,sql=TEXT,json=name]] [--force] [--dry-run]`

const usageMakeRepository = `Usage:
  gos make:repository <module> [--module=<module-path>] [--db=mysql] [--table=<table>] [--fields=name:string[:unique,nullable,size=320,default=value,sql=TEXT,json=name]] [--with-migration] [--migration-dir=migrations] [--register] [--force] [--dry-run]`

const usageMakeMigration = `Usage:
  gos make:migration <name> [--dir=migrations] [--force] [--dry-run]`

const usageMakeTest = `Usage:
  gos make:test <usecase|handler|repository> <name> [--module=<module-path>] [--force] [--dry-run]`

const usageMakeCommand = `Usage:
  gos make:command <name> [--module=<module-path>] [--register] [--force] [--dry-run]`
