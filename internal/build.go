package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/imdario/mergo"
	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl.cli/internal/iotools"
	"github.com/invopop/gobl/dsig"
	"github.com/invopop/gobl/schema"
)

// BuildOptions are the options to pass to the Build function.
type BuildOptions struct {
	Template   io.Reader
	Data       io.Reader
	Envelop    bool // when true, data is a document not envelope
	DocType    string
	SetYAML    map[string]string
	SetString  map[string]string
	SetFile    map[string]string
	PrivateKey *dsig.PrivateKey
}

// decodeInto unmarshals in as YAML, then merges it into dest.
func decodeInto(ctx context.Context, dest *map[string]interface{}, in io.Reader) error {
	var intermediate map[string]interface{}
	dec := yaml.NewDecoder(iotools.CancelableReader(ctx, in))
	if err := dec.Decode(&intermediate); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := mergo.Merge(dest, intermediate, mergo.WithOverride); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	return nil
}

// Build builds and validates a GOBL document from opts.
func Build(ctx context.Context, opts BuildOptions) (*gobl.Envelope, error) {
	values, err := parseSets(opts)
	if err != nil {
		return nil, err
	}
	var intermediate map[string]interface{}

	if opts.Template != nil {
		if err = decodeInto(ctx, &intermediate, opts.Template); err != nil {
			return nil, err
		}
	}
	if err = decodeInto(ctx, &intermediate, opts.Data); err != nil {
		return nil, err
	}

	if err = mergo.Merge(&intermediate, values, mergo.WithOverride); err != nil {
		return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	if opts.DocType != "" {
		schema := FindType(opts.DocType)
		if schema == "" {
			return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("unrecognized doc type: %q", opts.DocType))
		}
		if err = mergo.Merge(&intermediate, schemaDocumentData(opts, schema)); err != nil {
			return nil, err
		}
	}

	encoded, err := json.Marshal(intermediate)
	if err != nil {
		return nil, err
	}
	var env *gobl.Envelope
	if opts.Envelop {
		env = gobl.NewEnvelope()
		if err = json.Unmarshal(encoded, env.Document); err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	} else {
		env = new(gobl.Envelope)
		if err = json.Unmarshal(encoded, env); err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	if len(env.Signatures) > 0 {
		return nil, echo.NewHTTPError(http.StatusConflict, "document has already been signed")
	}

	if err = env.Complete(); err != nil {
		return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	if opts.PrivateKey == nil {
		return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, "signing key required")
	}
	if !env.Head.Draft {
		if err = env.Sign(opts.PrivateKey); err != nil {
			return nil, err
		}
	}
	return env, nil
}

func schemaDocumentData(opts BuildOptions, schema schema.ID) map[string]interface{} {
	doc := map[string]interface{}{
		"$schema": schema,
	}
	if opts.Envelop {
		return doc
	}
	return map[string]interface{}{
		"doc": doc,
	}
}

func parseSets(opts BuildOptions) (map[string]interface{}, error) {
	values := map[string]interface{}{}
	keys := make([]string, 0, len(opts.SetYAML))
	for k := range opts.SetYAML {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := opts.SetYAML[k]
		var parsed interface{}
		if err := yaml.Unmarshal([]byte(v), &parsed); err != nil {
			return nil, err
		}
		if err := setValue(&values, k, parsed); err != nil {
			return nil, err
		}
	}

	keys = make([]string, 0, len(opts.SetString))
	for k := range opts.SetString {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := opts.SetString[k]
		if err := setValue(&values, k, v); err != nil {
			return nil, err
		}
	}

	keys = make([]string, 0, len(opts.SetFile))
	for k := range opts.SetFile {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := opts.SetFile[k]
		f, err := os.Open(v)
		if err != nil {
			return nil, err
		}
		defer f.Close() // nolint:errcheck
		dec := yaml.NewDecoder(f)
		var val interface{}
		if err := dec.Decode(&val); err != nil {
			return nil, err
		}
		if err := setValue(&values, k, val); err != nil {
			return nil, err
		}
	}
	return values, nil
}

func setValue(values *map[string]interface{}, key string, value interface{}) error {
	key = strings.ReplaceAll(key, `\.`, "\x00")

	// If the key starts with '.', we treat that as the root of the
	// target object
	if key == "." {
		return mergo.Merge(values, value, mergo.WithOverride)
	}
	if len(key) > 1 && key[0] == '.' {
		key = key[1:]
	}

	for {
		i := strings.LastIndex(key, ".")
		if i == -1 {
			break
		}
		value = map[string]interface{}{
			strings.ReplaceAll(key[i+1:], "\x00", "."): value,
		}
		key = key[:i]
	}
	return mergo.Merge(values, map[string]interface{}{
		strings.ReplaceAll(key, "\x00", "."): value,
	}, mergo.WithOverride)
}
