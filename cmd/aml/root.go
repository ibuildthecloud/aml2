package main

import (
	"github.com/spf13/cobra"
)

type AML struct {
	Output string `usage:"output" short:"o" default:"json"`
}

func (a *AML) Customize(cmd *cobra.Command) {
	cmd.CompletionOptions.HiddenDefaultCmd = true
	cmd.SilenceUsage = true
	cmd.AddCommand(NewEval(a))
}

func (a *AML) Run(cmd *cobra.Command, args []string) error {
	return cmd.Usage()
}
