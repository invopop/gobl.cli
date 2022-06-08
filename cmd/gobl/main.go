// The gobl command provides a command-line interface to the GOBL library.

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"

	"github.com/invopop/gobl"
)

// Build Data
var (
	Version     = "dev"
	BuildDate   = ""
	BuildCommit = ""
)

func main() {
	if err := run(); err != nil {
		echoErr := new(echo.HTTPError)
		if errors.As(err, &echoErr) {
			msg := echoErr.Message
			int := echoErr.Internal
			switch {
			case msg != "" && int != nil:
				err = fmt.Errorf("%v: %w", msg, int)
			case int != nil:
				err = int
			default:
				err = fmt.Errorf("%v", msg)
			}
		}
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	return root().cmd().ExecuteContext(ctx)
}

func inputFilename(args []string) string {
	if len(args) > 0 && args[0] != "-" {
		return args[0]
	}
	return ""
}

func version() *cobra.Command {
	return &cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, _ []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "CLI: %s\nGOBL %s\n", Version, gobl.VERSION)
			if BuildCommit != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Commit: %s\n", BuildCommit)
			}
			if BuildDate != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Date: %s\n", BuildDate)
			}
		},
	}
}
