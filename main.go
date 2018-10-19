package main

import "github.com/bakaoh/lavato/cmd"

var revision = ""

func main() {
	cmd.SetRevision(revision)
	cmd.Execute()
}
