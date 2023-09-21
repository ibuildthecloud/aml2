package main

import (
	"github.com/acorn-io/cmd"
)

func main() {
	cmd.Main(cmd.Command(&AML{}))
}
