package internal

import (
	"context"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl/schema"
)

type BuildOptions struct {
	*ParseOptions
}

// Build builds and validates GOBL data. Only structured errors are returned,
// which is a break from regular Go convention and replicated on all the main
// internal CLI functions. The object is to ensure that errors are always
// structured in a consistent manner.
func Build(ctx context.Context, opts *BuildOptions) (any, error) {
	obj, err := parseGOBLData(ctx, opts.ParseOptions)
	if err != nil {
		return nil, wrapError(StatusUnprocessableEntity, err)
	}

	if env, ok := obj.(*gobl.Envelope); ok {
		// Signed documents should be regarded as immutable.
		// Attempting to build an already signed document returns an error.
		if len(env.Signatures) > 0 {
			return nil, wrapErrorf(StatusConflict, "document has already been signed")
		}

		if err := env.Calculate(); err != nil {
			return nil, wrapError(StatusUnprocessableEntity, err)
		}

		if err := env.Validate(); err != nil {
			return nil, wrapError(StatusUnprocessableEntity, err)
		}

		return env, nil
	}

	if doc, ok := obj.(*schema.Object); ok {
		if c, ok := doc.Instance().(schema.Calculable); ok {
			if err := c.Calculate(); err != nil {
				err = gobl.WrapError(err) // schema object errors need to be wrapped
				return nil, wrapError(StatusUnprocessableEntity, err)
			}
		}

		if err := doc.Validate(); err != nil {
			err = gobl.WrapError(err) // schema object errors need to be wrapped
			return nil, wrapError(StatusUnprocessableEntity, err)
		}

		return doc, nil
	}

	panic("parsed data must be either an envelope or a document")
}
