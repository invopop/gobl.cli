package internal

import (
	"reflect"
	"testing"

	"github.com/invopop/gobl/schema"
)

func TestFindType(t *testing.T) {
	const (
		idInvoice = "https://gobl.org/draft-0/bill/invoice"
	)
	type Invoice struct{}
	r := map[reflect.Type]schema.ID{
		reflect.TypeOf(Invoice{}): idInvoice,
	}

	t.Run("exact schema match", func(t *testing.T) {
		got := findType(r, idInvoice)
		if got != idInvoice {
			t.Errorf("Unexpected result: %v", got)
		}
	})
	t.Run("exact type match", func(t *testing.T) {
		got := findType(r, "Invoice")
		if got != idInvoice {
			t.Errorf("Unexpected result: %v", got)
		}
	})
	t.Run("exact type match with package", func(t *testing.T) {
		got := findType(r, "internal.Invoice")
		if got != idInvoice {
			t.Errorf("Unexpected result: %v", got)
		}
	})
	t.Run("wrong package", func(t *testing.T) {
		got := findType(r, "wrongpkg.Invoice")
		if got != "" {
			t.Errorf("Unexpected result: %v", got)
		}
	})
}
