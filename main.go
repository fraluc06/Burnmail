package main

import (
	"burnmail/cmd"
)

var Version = "1.4.1"

func main() {
	cmd.Version = Version
	cmd.Execute()
}
