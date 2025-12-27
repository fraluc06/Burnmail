package main

import (
	"burnmail/cmd"
)

var Version = "1.3.0"

func main() {
	cmd.Version = Version
	cmd.Execute()
}
