package main

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/invopop/gobl.cli/internal"
)

type validateOpts struct {
	*rootOpts

	// Command options
	use   string
	short string
}

func validate(root *rootOpts) *validateOpts {
	return &validateOpts{
		rootOpts: root,
		use:      "validate [infile] [outfile]",
		short:    "Validate checks if the input is a valid GOBL document",
	}
}

func (opts *validateOpts) cmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.MaximumNArgs(2),
		RunE:  opts.runE,
		Use:   opts.use,
		Short: opts.short,
	}

	return cmd
}

func (opts *validateOpts) runE(cmd *cobra.Command, args []string) error {
	ctx := commandContext(cmd)

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

	env, err := internal.Validate(ctx, input)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(out)
	if opts.indent {
		enc.SetIndent("", "\t")
	}

	return enc.Encode(env)
}
