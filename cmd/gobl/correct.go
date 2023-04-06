package main

import (
	"encoding/json"

	"github.com/invopop/gobl.cli/internal"
	"github.com/spf13/cobra"
)

type correctOpts struct {
	*rootOpts
	credit bool
	debit  bool
}

func correct(root *rootOpts) *correctOpts {
	return &correctOpts{
		rootOpts: root,
	}
}

func (o *correctOpts) cmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.MaximumNArgs(2),
		RunE:  o.runE,
		Use:   "correct [infile] [outfile]",
		Short: "Build a corrective document from the provided input",
	}

	f := cmd.Flags()
	f.BoolVarP(&o.credit, "credit", "", false, "Generate a credit note or negative corrective document.")
	f.BoolVarP(&o.debit, "debit", "", false, "Generate a debit note.")

	return cmd
}

func (o *correctOpts) runE(cmd *cobra.Command, args []string) error {
	ctx := commandContext(cmd)

	input, err := openInput(cmd, args)
	if err != nil {
		return err
	}
	defer input.Close() // nolint:errcheck

	out, err := o.openOutput(cmd, args)
	if err != nil {
		return err
	}
	defer out.Close() // nolint:errcheck

	cOpts := &internal.CorrectOptions{
		ParseOptions: &internal.ParseOptions{
			Data: input,
		},
		Credit: o.credit,
		Debit:  o.debit,
	}

	obj, err := internal.Correct(ctx, cOpts)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(out)
	if o.indent {
		enc.SetIndent("", "\t")
	}

	return enc.Encode(obj)
}
