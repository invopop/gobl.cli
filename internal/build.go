package internal

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/invopop/gobl"
)

// Build builds and validates GOBL data from build options, and transparently
// wraps a document in an envelope if needed.
func Build(ctx context.Context, opts ParseOptions) (*gobl.Envelope, error) {
	env, err := parseGOBLData(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Signed documents should be regarded as immutable.
	// Attempting to build an already signed document returns an error.
	if len(env.Signatures) > 0 {
		return nil, echo.NewHTTPError(http.StatusConflict, "document has already been signed")
	}

	if err := env.Calculate(); err != nil {
		return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}

	if err := env.Validate(); err != nil {
		return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}

	return env, nil
}
