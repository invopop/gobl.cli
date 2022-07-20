package internal

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/invopop/gobl"
	"github.com/labstack/echo/v4"
)

// Calculate parses a GOBL document and performs calculations, including invoice
// totals and envelope header digest.
func Calculate(ctx context.Context, opts *BuildOptions) (*gobl.Envelope, error) {
	encoded, err := prepareIntermediate(ctx, opts, docInEnvelopeSchemaData)
	if err != nil {
		return nil, err
	}

	// Prepare an empty envelope as we assume the consumer is providing one already.
	env := new(gobl.Envelope)
	if err := json.Unmarshal(encoded, env); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
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
