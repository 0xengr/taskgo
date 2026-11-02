// Command taskgo is a local-first task and project tracker for developers.
package main

import (
	"os"

	"github.com/0xengr/taskgo/internal/cli"
)

func main() {
	os.Exit(cli.Execute(os.Args[1:], os.Stdout, os.Stderr))
}
