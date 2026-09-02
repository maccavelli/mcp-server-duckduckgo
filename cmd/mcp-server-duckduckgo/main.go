package main

import (
	"fmt"
	"os"

	"github.com/maccavelli/mcplib/selfupdate"
)

func main() {
	err := Execute()
	if err == nil {
		return
	}
	// The library never exits the process. selfupdate.ExitCode returns 10 when
	// `update --check` found an actionable target and 1 for every other
	// failure, so an available update is scriptable rather than an error.
	code := selfupdate.ExitCode(selfupdate.Result{}, err)
	if code != 10 {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(code)
}
