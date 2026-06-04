package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jake/gola/internal/command"
)

func main() {
	if err := command.Execute(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
