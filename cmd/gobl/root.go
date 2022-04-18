package main

import "github.com/spf13/cobra"

type rootOpts struct {
	indent bool // when true, indent output, mainly for testing
}

func (o *rootOpts) setFlags(cmd *cobra.Command) {
	f := cmd.Flags()
	f.BoolVarP(&o.indent, "indent", "i", false, "format JSON output with indentation")
}
