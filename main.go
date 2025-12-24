package main

import (
	"burnmail/cmd"
)

var Version = "1.2.2"

func main() {
	cmd.Version = Version
	cmd.Execute()
}
