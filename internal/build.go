package internal

import (
	"context"
	"net/http"

	"github.com/invopop/gobl"
	"github.com/labstack/echo/v4"
)

type BuildOptions struct {
	*ParseOptions
	// When set to a non-nil value, the returned data is wrapped in an envelope (if needed)
	// with its `draft` property set to true or false.
	Draft *bool
}

// Build builds and validates GOBL data.
func Build(ctx context.Context, opts *BuildOptions) (interface{}, error) {
	obj, err := parseGOBLData(ctx, opts.ParseOptions)
	if err != nil {
		return nil, err
	}

	if env, ok := obj.(*gobl.Envelope); ok {
		// Signed documents should be regarded as immutable.
		// Attempting to build an already signed document returns an error.
		if len(env.Signatures) > 0 {
			return nil, echo.NewHTTPError(http.StatusConflict, "document has already been signed")
		}

		if opts.Draft != nil {
			env.Head.Draft = *opts.Draft
		}

		if err := env.Calculate(); err != nil {
			return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}

		if err := env.Validate(); err != nil {
			return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}

		return env, nil
	}

	if doc, ok := obj.(*gobl.Document); ok {
		if opts.Draft != nil {
			return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, "cannot set draft status on non-envelope document")
		}
		if c, ok := doc.Instance().(gobl.Calculable); ok {
			if err := c.Calculate(); err != nil {
				return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
			}
		}

		if err := doc.Validate(); err != nil {
			return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}

		return doc, nil
	}

	panic("parsed data must be either an envelope or a document")
}
