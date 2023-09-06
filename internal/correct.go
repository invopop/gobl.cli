package internal

import (
	"context"
	"net/http"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cal"
	"github.com/invopop/gobl/schema"
	"github.com/labstack/echo/v4"
)

// CorrectOptions define all the basic options required to build a corrective
// document from the input.
type CorrectOptions struct {
	// we don't need all of the parse options
	*ParseOptions
	Credit bool
	Debit  bool
	Date   cal.Date
	Data   []byte // raw json of correction options
}

// Correct takes a base document as input and builds a corrective document
// for the output using the base document for input.
func Correct(ctx context.Context, opts *CorrectOptions) (interface{}, error) {
	obj, err := parseGOBLData(ctx, opts.ParseOptions)
	if err != nil {
		return nil, err
	}

	eopts := make([]schema.Option, 0)
	if len(opts.Data) > 0 {
		eopts = append(eopts, bill.WithData(opts.Data))
	}
	if opts.Credit {
		eopts = append(eopts, bill.Credit)
	}
	if opts.Debit {
		eopts = append(eopts, bill.Debit)
	}
	if !opts.Date.IsZero() {
		eopts = append(eopts, bill.WithIssueDate(opts.Date))
	}

	if env, ok := obj.(*gobl.Envelope); ok {
		e2, err := env.Correct(eopts...)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}
		if err = e2.Validate(); err != nil {
			return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}
		return e2, nil
	}

	if doc, ok := obj.(*schema.Object); ok {
		// Documents are updated in place
		if err := doc.Correct(eopts...); err != nil {
			return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}
		if err = doc.Validate(); err != nil {
			return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}
		return doc, nil
	}

	panic("input must be either an envelope or a document")
}
