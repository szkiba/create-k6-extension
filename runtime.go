package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/mgutz/ansi"
)

type runtime struct {
	*terminal.Stdio
	args     []string
	exit     func(int)
	lookPath func(string) (string, error)
}

//nolint:forbidigo
func stdRuntime() *runtime {
	return &runtime{
		Stdio: &terminal.Stdio{
			In:  os.Stdin,
			Out: os.Stdout,
			Err: os.Stderr,
		},
		args:     os.Args,
		exit:     os.Exit,
		lookPath: exec.LookPath,
	}
}

func (rt *runtime) fail(err error) {
	fmt.Fprintf(rt.Err, "error: %s\n", err.Error())
	rt.exit(1)
}

func (rt *runtime) prerequisite() {
	rt.require(
		"go",
		"To use this command, you need the go toolchain!",
		"https://go.dev/doc/install",
	)

	rt.require(
		"git",
		"To use this command, you need the git CLI!",
		"https://git-scm.com/downloads",
	)
}

func (rt *runtime) require(cmd, msg, link string) {
	if _, err := rt.lookPath("git"); err == nil {
		return
	}

	fmt.Fprintf(
		terminal.NewAnsiStderr(rt.Out),
		"%s\n%s\n",
		ansi.Color(msg, "red"),
		ansi.Color(link, "cyan"),
	)

	rt.fail(fmt.Errorf("%w: %s", errPrerequisite, cmd))
}

var errPrerequisite = errors.New("missing prerequisite")
