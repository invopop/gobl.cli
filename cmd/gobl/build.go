package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/invopop/gobl.cli/internal"
)

var (
	boolTrue  = true
	boolFalse = false
)

type buildOpts struct {
	*rootOpts
	set        map[string]string
	setFiles   map[string]string
	setStrings map[string]string
	template   string
	docType    string
	envelop    bool
	draft      bool
	notDraft   bool

	// Command options
	use   string
	short string
}

func build(root *rootOpts) *buildOpts {
	return &buildOpts{
		rootOpts: root,
		use:      "build [infile] [outfile]",
		short:    "Calculate and validate a document, wrapping it in envelope if needed",
	}
}

func (b *buildOpts) cmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.MaximumNArgs(2),
		RunE:  b.runE,
		Use:   b.use,
		Short: b.short,
	}

	f := cmd.Flags()
	f.StringToStringVar(&b.set, "set", nil, "Set value from the command line")
	f.StringToStringVar(&b.setFiles, "set-file", nil, "Set value from the specified YAML or JSON file")
	f.StringToStringVar(&b.setStrings, "set-string", nil, "Set STRING value from the command line")
	f.StringVarP(&b.template, "template", "T", "", "Template YAML/JSON file into which data is merged")
	f.StringVarP(&b.docType, "type", "t", "", "Specify the document type")
	f.BoolVarP(&b.envelop, "envelop", "e", false, "format JSON output with indentation")
	f.BoolVarP(&b.draft, "draft", "d", false, "Set envelope as draft")
	f.BoolVarP(&b.notDraft, "not-draft", "n", false, "Set envelope as non-draft")

	return cmd
}

func (b *buildOpts) runE(cmd *cobra.Command, args []string) error {
	ctx := commandContext(cmd)

	var template io.Reader
	if b.template != "" {
		f, err := os.Open(b.template)
		if err != nil {
			return err
		}
		defer f.Close() // nolint:errcheck
		template = f
	}

	input, err := openInput(cmd, args)
	if err != nil {
		return err
	}
	defer input.Close() // nolint:errcheck

	out, err := b.openOutput(cmd, args)
	if err != nil {
		return err
	}
	defer out.Close() // nolint:errcheck

	buildOpts := &internal.BuildOptions{
		ParseOptions: &internal.ParseOptions{
			Template:  template,
			Data:      input,
			SetFile:   b.setFiles,
			SetYAML:   b.set,
			SetString: b.setStrings,
			DocType:   b.docType,
			Envelop:   b.envelop,
		},
	}

	switch {
	case b.draft && b.notDraft:
		return errors.New("draft and not-draft cannot both be set")
	case b.draft:
		buildOpts.Draft = &boolTrue
	case b.notDraft:
		buildOpts.Draft = &boolFalse
	}

	env, err := internal.Build(ctx, buildOpts)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(out)
	if b.indent {
		enc.SetIndent("", "\t") // Removing JSON formatting by default
	}
	return enc.Encode(env)
}
