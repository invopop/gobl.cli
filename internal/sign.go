package internal

import (
	"context"
	"net/http"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl/dsig"
	"github.com/labstack/echo/v4"
)

// SignOptions are the options used for signing a GOBL document.
type SignOptions struct {
	ParseOptions
	PrivateKey *dsig.PrivateKey
}

// Sign parses a GOBL document into an envelope, performs calculations,
// validates it, and finally signs its headers. The parsed envelope *must* be a
// draft, or else an error is returned.
func Sign(ctx context.Context, opts SignOptions) (*gobl.Envelope, error) {
	// Always envelop incoming data.
	opts.Envelop = true

	obj, err := parseGOBLData(ctx, opts.ParseOptions)
	if err != nil {
		return nil, err
	}

	env, ok := obj.(*gobl.Envelope)
	if !ok {
		panic("parsed sign data must be an envelope")
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
