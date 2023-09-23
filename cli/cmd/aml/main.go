package main

import (
	"github.com/acorn-io/aml/cli/pkg/cmds"
	"github.com/acorn-io/cmd"
)

func main() {
	cmd.Main(cmd.Command(&cmds.AML{}))
}
