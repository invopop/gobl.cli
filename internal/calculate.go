package internal

import (
	"context"
	"net/http"

	"github.com/invopop/gobl"
	"github.com/labstack/echo/v4"
)

// Calculate parses a GOBL document and performs calculations, including invoice
// totals and envelope header digest.
func Calculate(ctx context.Context, opts *BuildOptions) (*gobl.Envelope, error) {
	env, err := parseBuildData(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Signed documents should be regarded as immutable.
	// Attempting to calculate an already signed document returns an error.
	if len(env.Signatures) > 0 {
		return nil, echo.NewHTTPError(http.StatusConflict, "document has already been signed")
	}

	if err := env.Calculate(); err != nil {
		return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}

	return env, nil
}
