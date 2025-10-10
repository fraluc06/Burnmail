package main

import (
	"burnmail/cmd"
)

var Version = "0.1.0"

func main() {
	cmd.Version = Version
	cmd.Execute()
}
