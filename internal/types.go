package internal

import (
	"path/filepath"
	"reflect"

	"github.com/invopop/gobl/schema"
)

func FindType(term string) schema.ID {
	return findType(schema.Types(), term)
}

func findType(types map[reflect.Type]schema.ID, term string) schema.ID {
	for typ, id := range types {
		if term == string(id) {
			return id
		}
		if term == typ.Name() {
			return id
		}
		if term == filepath.Base(typ.PkgPath())+"."+typ.Name() {
			return id
		}
	}
	return ""
}
