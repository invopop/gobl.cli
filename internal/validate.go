package internal

import (
	"context"
	"io"
	"net/http"

	"github.com/invopop/gobl"
	"github.com/labstack/echo/v4"
)

// Validate asserts the contents of the envelope and document are correct.
func Validate(ctx context.Context, r io.Reader) (*gobl.Envelope, error) {
	opts := ParseOptions{
		Data: r,
	}
	env, err := parseGOBLData(ctx, opts)
	if err != nil {
		return nil, err
	}

	if err := env.Validate(); err != nil {
		return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}

	return env, nil
}
