// main.go
package main

import (
	"os"
	"ssl-tools/cmd"
	tuicmd "ssl-tools/cmd/ssl-tools-tui"
)

func main() {
	// if no args passed start BubbleTea TUI
	if len(os.Args) == 1 {
		tuicmd.RunTUI()
		return
	}
	cmd.Execute()
}
