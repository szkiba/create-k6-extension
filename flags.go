package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/pflag"
	"golang.org/x/term"
)

func usage(out io.Writer, flags *pflag.FlagSet) {
	fmt.Fprintf(out,
		"Usage: %s [flags] [directory]\n\nFlags:\n%s",
		_appname,
		flags.FlagUsages(),
	)
}

func flagset(opts *options, terminal bool) *pflag.FlagSet {
	flags := pflag.NewFlagSet(_appname, pflag.ContinueOnError)

	flags.BoolVar(&opts.NoAsk, "no-ask", !terminal, "disable interactive questions")
	flags.StringVar(&opts.Name, "name", "", "extension name")
	flags.StringVar(&opts.Summary, "summary", "", "a brief summary of the extension")
	flags.StringVar(&opts.GoModule, "go-module", "", "go module path")
	flags.StringVar(&opts.GoPackage, "go-package", "", "go package name (default: extension name)")

	flags.StringVar(&opts.GitOrigin, "git-origin", "", "git origin URL")

	flags.StringVar(&opts.RepoOwner, "repo-owner", "", "GitHub repository owner")
	flags.StringVar(&opts.RepoName, "repo-name", "", "GitHub repository name")
	flags.StringVar(
		&opts.RepoProtocol,
		"repo-protocol",
		"ssh",
		"git repository origin protocol (ssh or https)",
	)

	flags.BoolVar(&opts.NoGitInit, "no-git-init", false, "disable git module initialization")
	flags.BoolVar(&opts.NoGitOrigin, "no-git-origin", false, "disable setting git origin")

	flags.BoolVar(&opts.debug, "debug", false, "enable debug output")

	return flags
}

func getopts(rt *runtime) (*options, error) {
	opts := new(options)
	flags := flagset(opts, term.IsTerminal(int(rt.In.Fd())))

	kindstr := flags.String("type", string(javascript), "extension type (JavaScript or Output)")
	ver := flags.Bool("version", false, "print version")
	help := flags.BoolP("help", "h", false, "print this help message")

	if err := flags.Parse(rt.args); err != nil {
		return nil, err
	}

	opts.Kind = kind(*kindstr)

	if *help {
		usage(rt.Err, flags)

		return nil, pflag.ErrHelp
	}

	if *ver {
		fmt.Fprintf(rt.Err, "%s version %s\n", _appname, _version)

		return nil, pflag.ErrHelp
	}

	if flags.NArg() > 2 {
		return nil, errTooManyArg
	}

	if flags.NArg() == 2 {
		opts.Dir = flags.Arg(1)
	}

	opts.update()

	_, err := rt.lookPath("xk6")
	opts.installed = err == nil

	if !opts.NoAsk {
		return opts, nil
	}

	opts.guess()

	if len(opts.Name) == 0 {
		return nil, fmt.Errorf("%w: %s", errMissingFlag, "name")
	}

	if len(opts.GoModule) == 0 {
		return nil, fmt.Errorf("%w: %s", errMissingFlag, "go-module")
	}

	if len(opts.Dir) == 0 {
		return nil, fmt.Errorf("%w: %s", errMissingArg, "directory")
	}

	return opts, nil
}

var (
	errMissingFlag = errors.New("missing required flag")
	errTooManyArg  = errors.New("too many arguments")
	errMissingArg  = errors.New("missing argument")
)
