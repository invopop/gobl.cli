package main

import (
	"encoding/json"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/invopop/gobl.cli/internal"
)

type calculateOpts struct {
	*rootOpts
	set        map[string]string
	setFiles   map[string]string
	setStrings map[string]string
	template   string
	docType    string

	// Command options
	use   string
	short string
}

func calculate(root *rootOpts) *calculateOpts {
	return &calculateOpts{
		rootOpts: root,
		use:      "calculate [infile] [outfile]",
		short:    "Calculate performs calculations on a document",
	}
}

func (opts *calculateOpts) cmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.MaximumNArgs(2),
		RunE:  opts.runE,
		Use:   opts.use,
		Short: opts.short,
	}

	f := cmd.Flags()
	f.StringToStringVar(&opts.set, "set", nil, "Set value from the command line")
	f.StringToStringVar(&opts.setFiles, "set-file", nil, "Set value from the specified YAML or JSON file")
	f.StringToStringVar(&opts.setStrings, "set-string", nil, "Set STRING value from the command line")
	f.StringVarP(&opts.template, "template", "T", "", "Template YAML/JSON file into which data is merged")
	f.StringVarP(&opts.docType, "type", "t", "", "Specify the document type")

	return cmd
}

func (opts *calculateOpts) runE(cmd *cobra.Command, args []string) error {
	ctx := commandContext(cmd)

	var template io.Reader
	if opts.template != "" {
		f, err := os.Open(opts.template)
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

	out, err := opts.openOutput(cmd, args)
	if err != nil {
		return err
	}
	defer out.Close() // nolint:errcheck

	parseOpts := internal.ParseOptions{
		Template:  template,
		Data:      input,
		SetFile:   opts.setFiles,
		SetYAML:   opts.set,
		SetString: opts.setStrings,
		DocType:   opts.docType,
	}

	env, err := internal.Calculate(ctx, parseOpts)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(out)
	if opts.indent {
		enc.SetIndent("", "\t")
	}

	return enc.Encode(env)
}
