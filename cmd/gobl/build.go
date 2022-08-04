package main

import (
	"encoding/json"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/invopop/gobl.cli/internal"
	"github.com/invopop/gobl/dsig"
)

type buildOpts struct {
	*rootOpts
	set            map[string]string
	setFiles       map[string]string
	setStrings     map[string]string
	template       string
	privateKeyFile string
	docType        string

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

	pkFilename, err := expandHome(b.privateKeyFile)
	if err != nil {
		return err
	}
	keyFile, err := os.Open(pkFilename)
	if err != nil {
		return err
	}
	defer keyFile.Close() // nolint:errcheck

	key := new(dsig.PrivateKey)
	if err = json.NewDecoder(keyFile).Decode(key); err != nil {
		return err
	}

	parseOpts := internal.ParseOptions{
		Template:  template,
		Data:      input,
		SetFile:   b.setFiles,
		SetYAML:   b.set,
		SetString: b.setStrings,
		DocType:   b.docType,
	}

	env, err := internal.Build(ctx, parseOpts)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(out)
	if b.indent {
		enc.SetIndent("", "\t") // Removing JSON formatting by default
	}
	return enc.Encode(env)
}
