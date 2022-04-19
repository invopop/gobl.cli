package main

import "github.com/spf13/cobra"

type rootOpts struct {
	indent              bool // when true, indent output, mainly for testing
	overwriteOutputFile bool
	inPlace             bool
}

func (o *rootOpts) setFlags(cmd *cobra.Command) {
	f := cmd.PersistentFlags()
	f.BoolVarP(&o.indent, "indent", "i", false, "format JSON output with indentation")
	f.BoolVarP(&o.overwriteOutputFile, "force", "f", false, "force writing output file, even if it exists")
	f.BoolVarP(&o.inPlace, "in-place", "w", false, "overwrite the input file in place  (only outputs JSON)")
}
