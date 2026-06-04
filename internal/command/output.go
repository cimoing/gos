package command

import (
	"fmt"
	"io"

	"github.com/jake/gola/internal/filesystem"
)

func printResult(stdout io.Writer, result *filesystem.Result) {
	for _, file := range result.Created {
		fmt.Fprintf(stdout, "  create %s\n", file)
	}
	for _, file := range result.Updated {
		fmt.Fprintf(stdout, "  update %s\n", file)
	}
	for _, file := range result.Skipped {
		fmt.Fprintf(stdout, "  skip %s\n", file)
	}
}
