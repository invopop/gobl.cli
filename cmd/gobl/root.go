package main

import "github.com/spf13/cobra"

type rootOpts struct {
	indent              bool // when true, indent output, mainly for testing
	overwriteOutputFile bool
	inPlace             bool
}

func root() *cobra.Command {
	opts := &rootOpts{}
	cmd := &cobra.Command{
		Use:           "gobl",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	opts.setFlags(cmd)

	cmd.AddCommand(verify().cmd())
	cmd.AddCommand(envelop().cmd()) // see build
	cmd.AddCommand(build(opts).cmd())
	cmd.AddCommand(version())
	cmd.AddCommand(serve().cmd())
	cmd.AddCommand(keygen(opts).cmd())
	return cmd
}

func (o *rootOpts) setFlags(cmd *cobra.Command) {
	f := cmd.PersistentFlags()
	f.BoolVarP(&o.indent, "indent", "i", false, "format JSON output with indentation")
	f.BoolVarP(&o.overwriteOutputFile, "force", "f", false, "force writing output file, even if it exists")
	f.BoolVarP(&o.inPlace, "in-place", "w", false, "overwrite the input file in place  (only outputs JSON)")
}

func (o *rootOpts) outputFilename(args []string) string {
	if o.inPlace {
		return inputFilename(args)
	}
	if len(args) >= 2 && args[1] != "-" {
		return args[1]
	}
	return ""
}
