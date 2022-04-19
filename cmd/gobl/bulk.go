package main

import "github.com/spf13/cobra"

type bulkOpts struct {
	*rootOpts
}

func (o *bulkOpts) runE(cmd *cobra.Command, args []string) error {
	return nil
}
