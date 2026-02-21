package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	rootcmd "github.com/go-extras/godexer/cmd/godexer/root"
	"github.com/go-extras/godexer/cmd/godexer/shared"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		var exitErr *shared.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	return rootcmd.New().ExecuteContext(ctx)
}
