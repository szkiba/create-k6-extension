// Package main contains the create-k6-extension CLI tool.
package main

import (
	"errors"

	"github.com/spf13/pflag"
)

var (
	_appname = "create-k6-extension" //nolint:gochecknoglobals
	_version = "dev"                 //nolint:gochecknoglobals
)

func main() {
	rt := stdRuntime()

	rt.prerequisite()

	run(rt)
}

func run(rt *runtime) {
	var err error
	var opts *options

	opts, err = getopts(rt)
	if errors.Is(err, pflag.ErrHelp) {
		return
	}

	if err != nil {
		rt.fail(err)
	}

	var confirm bool

	confirm, err = ask(opts, rt.Stdio)
	if err != nil {
		rt.fail(err)
	}

	if !confirm {
		return
	}

	err = create(opts, rt.Stdio)
	if err != nil {
		rt.fail(err)
	}
}
