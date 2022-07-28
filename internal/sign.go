package internal

import (
	"context"
	"net/http"

	"github.com/invopop/gobl"
	"github.com/labstack/echo/v4"
)

// Sign parses a GOBL document into an envelope, performs calculations,
// validates it, and finally signs its headers. The parsed envelope *must* be a
// draft, or else an error is returned.
func Sign(ctx context.Context, opts *BuildOptions) (*gobl.Envelope, error) {
	// TODO: `BuildOptions` should probably be renamed to `ParseOptions`,
	// as parsing a GOBL data seems to be the only (shared) purpose across
	// the principal CLI commands in this package.
	env, err := parseBuildData(ctx, opts)
	if err != nil {
		return nil, err
	}

	// A draft automatically becomes a non-draft (i.e. "final") document, This
	// way it's possible to sign a document in draft state with a single
	// command.
	if env.Head != nil && env.Head.Draft {
		env.Head.Draft = false
	}

	if err := env.Calculate(); err != nil {
		return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}

	// Sign envelope headers. Validation is done transparently in `Sign`.
	if err := env.Sign(opts.PrivateKey); err != nil {
		return nil, err
	}

	return env, nil
}
