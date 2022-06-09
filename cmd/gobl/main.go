// The gobl command provides a command-line interface to the GOBL library.

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"

	"github.com/invopop/gobl"
)

// build data provided by goreleaser and mage setup
var (
	version = "dev"
	date    = ""
)

var versionOutput = struct {
	Version string `json:"version"`
	GOBL    string `json:"gobl"`
	Date    string `json:"date,omitempty"`
}{
	Version: version,
	GOBL:    string(gobl.VERSION),
	Date:    date,
}

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

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use: "version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "\t") // always indent version
			return enc.Encode(versionOutput)
		},
	}
}
