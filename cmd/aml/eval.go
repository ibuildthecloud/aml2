package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/acorn-io/aml"
	"github.com/acorn-io/cmd"
	"github.com/spf13/cobra"
)

type Eval struct {
	aml *AML
}

func NewEval(aml *AML) *cobra.Command {
	return cmd.Command(&Eval{
		aml: aml,
	})
}

func (e *Eval) Customize(cmd *cobra.Command) {
	cmd.Use = "eval FILE"
	cmd.Short = "Evaluate a file and output the result"
	cmd.Args = cobra.ExactArgs(1)
}

func (e *Eval) Run(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}

	out := map[string]any{}
	err = aml.Unmarshal(data, &out, aml.Option{
		SourceName: args[0],
	})
	if err != nil {
		return err
	}

	data, err = json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}
