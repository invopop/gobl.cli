package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl.cli/internal"
	"github.com/invopop/gobl/dsig"
)

const (
	defaultHTTPPort = 80

	// If you customize this server, you should change this.
	vendorName = "Invopop Ltd."
)

type serveOpts struct {
	httpPort int
}

func serve() *serveOpts {
	return &serveOpts{}
}

func (s *serveOpts) cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Launch an HTTP server",
		RunE:  s.runE,
	}
	f := cmd.Flags()

	f.IntVarP(&s.httpPort, "port", "p", defaultHTTPPort, "HTTP port to listen on")

	return cmd
}

func (s *serveOpts) runE(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	e := echo.New()

	e.GET("/", s.version)
	e.POST("/build", s.build)
	e.POST("/verify", s.verify)
	e.POST("/key", s.keygen)
	e.POST("/bulk", s.bulk)

	var startErr error
	go func() {
		err := e.Start(":" + strconv.Itoa(s.httpPort))
		if !errors.Is(err, http.ErrServerClosed) {
			startErr = err
		}
		cancel() // Ensure we stop the context if we have a startup error
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return startErr
}

func (s *serveOpts) version(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"gobl":    "Welcome",
		"version": gobl.VERSION,
		"vendor": map[string]interface{}{
			"name": vendorName,
		},
	})
}

func (s *serveOpts) build(c echo.Context) error {
	opts, err := prepareBuildOpts(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	env, err := internal.Build(ctx, *opts)
	if err != nil {
		return err
	}

	blob, err := marshal(c)(env)
	if err != nil {
		return err
	}

	return c.JSONBlob(http.StatusOK, blob)
}

func prepareBuildOpts(c echo.Context) (*internal.BuildOptions, error) {
	ct, _, _ := mime.ParseMediaType(c.Request().Header.Get("Content-Type"))
	if ct != "application/json" {
		return nil, echo.NewHTTPError(http.StatusUnsupportedMediaType)
	}
	req := new(internal.BuildRequest)
	if err := c.Bind(req); err != nil {
		return nil, err
	}
	if len(req.Data) == 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "no payload")
	}
	opts := &internal.BuildOptions{
		ParseOptions: internal.ParseOptions{
			Data:    bytes.NewReader(req.Data),
			DocType: req.DocType,
		},
		Draft: req.Draft,
	}
	if len(req.Template) != 0 {
		opts.Template = bytes.NewReader(req.Template)
	}
	return opts, nil
}

func (s *serveOpts) verify(c echo.Context) error {
	ct, _, _ := mime.ParseMediaType(c.Request().Header.Get("Content-Type"))
	if ct != "application/json" {
		return echo.NewHTTPError(http.StatusUnsupportedMediaType)
	}
	req := new(internal.VerifyRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	if err := internal.Verify(c.Request().Context(), bytes.NewReader(req.Data), req.PublicKey); err != nil {
		return err
	}
	blob, err := marshal(c)(&internal.VerifyResponse{OK: true})
	if err != nil {
		return err
	}

	return c.JSONBlob(http.StatusOK, blob)
}

func (s *serveOpts) keygen(c echo.Context) error {
	key := dsig.NewES256Key()

	blob, err := marshal(c)(internal.KeygenResponse{
		Private: key,
		Public:  key.Public(),
	})
	if err != nil {
		return err
	}

	return c.JSONBlob(http.StatusOK, blob)
}

func (s *serveOpts) bulk(c echo.Context) error {
	ctx := c.Request().Context()
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)

	enc := json.NewEncoder(c.Response())
	if c.QueryParam("indent") == "true" {
		enc.SetIndent("", "\t")
	}
	for result := range internal.Bulk(ctx, c.Request().Body) {
		if err := enc.Encode(result); err != nil {
			return err
		}
	}
	return nil
}
