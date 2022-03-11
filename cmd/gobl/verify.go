package main

import (
	"github.com/invopop/gobl.cli/internal"
	"github.com/spf13/cobra"
)

type verifyOpts struct{}

func verify() *verifyOpts {
	return &verifyOpts{}
}

func (v *verifyOpts) cmd() *cobra.Command {
	return &cobra.Command{
		Use:  "verify [infile]",
		Args: cobra.MaximumNArgs(1),
		RunE: v.runE,
	}
}

func (v *verifyOpts) runE(cmd *cobra.Command, args []string) error {
	input, err := openInput(cmd, args)
	if err != nil {
		return err
	}
	defer input.Close() // nolint:errcheck

	return internal.Verify(cmdContext(cmd), input, nil)
}
