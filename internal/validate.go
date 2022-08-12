package internal

import (
	"context"
	"io"
	"net/http"

	"github.com/invopop/gobl"
	"github.com/labstack/echo/v4"
)

// Validate asserts the contents of the envelope and document are correct.
func Validate(ctx context.Context, r io.Reader) error {
	opts := &ParseOptions{
		Data: r,
	}
	obj, err := parseGOBLData(ctx, opts)
	if err != nil {
		return err
	}

	if env, ok := obj.(*gobl.Envelope); ok {
		if err := env.Validate(); err != nil {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}
		return nil
	}

	if doc, ok := obj.(*gobl.Document); ok {
		if err := doc.Validate(); err != nil {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}
		return nil
	}

	return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid document type")
}
