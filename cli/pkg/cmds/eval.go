package cmds

import (
	"errors"
	"os"

	"github.com/acorn-io/aml"
	"github.com/acorn-io/aml/cli/pkg/flagargs"
	"github.com/acorn-io/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Eval struct {
	aml *AML

	ArgsFile string `usage:"Default arguments to pass" default:".args.acorn"`
}

func NewEval(aml *AML) *cobra.Command {
	return cmd.Command(&Eval{aml: aml}, cobra.Command{
		Use:   "eval [flags] FILE",
		Short: "Evaluate a file and output the result",
		Args:  cobra.MinimumNArgs(1),
	})
}

func (e *Eval) Customize(cmd *cobra.Command) {
	cmd.Flags().SetInterspersed(false)
}

func (e *Eval) Run(cmd *cobra.Command, args []string) error {
	filename := args[0]
	args = args[1:]

	argsData, profiles, err := flagargs.ParseArgs(e.ArgsFile, filename, args)
	if errors.Is(err, pflag.ErrHelp) {
		return nil
	} else if err != nil {
		return err
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	out := map[string]any{}
	err = aml.Unmarshal(data, &out, aml.Option{
		SourceName: filename,
		Args:       argsData,
		Profiles:   profiles,
		Context:    cmd.Context(),
	})
	if err != nil {
		return err
	}

	return e.aml.Output(out)
}
