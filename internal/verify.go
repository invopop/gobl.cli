package internal

import (
	"context"
	"io"
	"net/http"

	jsonyaml "github.com/invopop/yaml"
	"github.com/labstack/echo/v4"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl.cli/internal/iotools"
	"github.com/invopop/gobl/dsig"
)

// Verify reads a GOBL document from in, and returns an error if there are any
// validation errors.
func Verify(ctx context.Context, in io.Reader, key *dsig.PublicKey) error {
	body, err := io.ReadAll(iotools.CancelableReader(ctx, in))
	if err != nil {
		return err
	}
	env := new(gobl.Envelope)
	if err := jsonyaml.Unmarshal(body, env); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := env.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	if key == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "public key required")
	}
	if env.Head.Draft {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "document is a draft")
	}
	if err := env.Signatures[0].VerifyPayload(key, env); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	return nil
}
