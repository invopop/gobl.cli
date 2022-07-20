package internal

import (
	"context"
	"encoding/json"
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
	encoded, err := prepareIntermediate(ctx, opts, docInEnvelopeSchemaData)
	if err != nil {
		return nil, err
	}

	env := new(gobl.Envelope)
	if err := json.Unmarshal(encoded, env); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Check early if the envelope is a draft, to prevent possible unnecessary
	// calculation.
	if env.Head != nil && env.Head.Draft {
		return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, "cannot sign draft envelope")
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
