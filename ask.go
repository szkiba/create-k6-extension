package main

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/go-playground/validator/v10"
	"github.com/mgutz/ansi"
)

func ask(opts *options, stdio *terminal.Stdio) (bool, error) {
	a := newAsker(opts, stdio)

	return a.askLoop()
}

type asker struct {
	*terminal.Stdio
	opts     *options
	validate *validator.Validate
}

func newAsker(opts *options, stdio *terminal.Stdio) *asker {
	return &asker{
		Stdio:    stdio,
		opts:     opts,
		validate: validator.New(),
	}
}

func (a *asker) print(format string, args ...any) {
	fmt.Fprintf(a.Out, format, args...)
}

func (a *asker) askKind() error {
	a.opts.guessKind()

	//nolint:lll
	prompt := &survey.Select{
		Message: "Extension type:",
		Options: []string{string(javascript), string(output)},
		Help:    "k6 supports two ways to extend its native functionality. Select JavaScript to extend the JavaScript APIs available to your test scripts. Select Output to send metrics to a custom file format or service.",
		Default: string(a.opts.Kind),
	}

	return a.ask(&a.opts.Kind, prompt)
}

func (a *asker) askName() error {
	a.opts.guessName()

	var help string

	if a.opts.Kind == javascript {
		help = "The part of the JavaScript module name after the k6/x/ prefix."
	} else {
		help = "The name to pass to the k6 run --out flag."
	}

	return a.ask(
		&a.opts.Name,
		&survey.Input{
			Message: "Extension name:",
			Help:    help,
			Default: a.opts.Name,
		},
		a.expect(
			"alphanum,min=3,max=32,lowercase",
			"Minimum 3 maximum 32 lower case alphanumeric characters are required",
		),
	)
}

func (a *asker) askDir() error {
	a.opts.guessDir()

	return a.ask(
		&a.opts.Dir,
		&survey.Input{
			Message: "Directory name:",
			Help:    "The name of the output repository.",
			Default: a.opts.Dir,
		},
		survey.Required,
	)
}

func (a *asker) askSummary() error {
	return a.ask(
		&a.opts.Summary,
		&survey.Input{
			Message: "Short description:",
			Help:    "A brief summary of the extension. Try to describe the purpose of the extension in one sentence.",
			Default: a.opts.Summary,
		},
	)
}

func (a *asker) askConfirm() (bool, error) {
	prompt := &survey.Confirm{
		Message: "Are the above answers correct?",
		Help:    "Select `n` if you want to modify the answers. Press `ctrl-c` to exit.",
	}

	var ok bool

	return ok, survey.AskOne(prompt, &ok, survey.WithStdio(a.In, a.Out, a.Err))
}

func (a *asker) askNoGitInit() error {
	return a.ask(
		&a.opts.NoGitInit,
		&survey.Confirm{
			Message: "Disable git repository initialization:",
			Default: a.opts.NoGitInit,
			Help:    "By default, a git repository is created from the extension's directory with the `git init` command.",
		},
	)
}

func (a *asker) askNoGitOrigin() error {
	if a.opts.NoGitInit {
		return nil
	}

	return a.ask(
		&a.opts.NoGitOrigin,
		&survey.Confirm{
			Message: "Disable setting git origin:",
			Default: a.opts.NoGitOrigin,
			Help:    "By default, the origin git repository is set using the `git remote add` command.",
		},
	)
}

func (a *asker) askUseGitHub() error {
	a.opts.guessUseGitHub()

	if a.opts.NoGitInit {
		return nil
	}

	//nolint:lll
	return a.ask(
		&a.opts.UseGitHub,
		&survey.Confirm{
			Message: "Host the repository on GitHub:",
			Default: a.opts.UseGitHub,
			Help:    "The easiest way to host your extension's repository is to use GitHub. Choose `n` if you want to host the repository elsewhere.",
		},
	)
}

func (a *asker) askRepoOwner() error {
	a.opts.guessUseGitHub()

	if !a.opts.UseGitHub {
		return nil
	}

	return a.ask(
		&a.opts.RepoOwner,
		&survey.Input{
			Message: "GitHub repository owner:",
			Help:    "GitHub user or organization that owns the repository.",
			Default: a.opts.RepoOwner,
		},
		survey.Required,
	)
}

func (a *asker) askGitOrigin() error {
	a.opts.guessGitOrigin()

	if a.opts.NoGitInit || a.opts.NoGitOrigin {
		return nil
	}

	return a.ask(
		&a.opts.GitOrigin,
		&survey.Input{
			Message: "git origin URL:",
			Help:    "The git origin URL as you want to use it in the `git remote add` command.",
			Default: a.opts.GitOrigin,
		},
		survey.Required,
	)
}

func (a *asker) askRepoName() error {
	a.opts.guessUseGitHub()
	a.opts.guessRepoName()

	if !a.opts.UseGitHub {
		return nil
	}

	prefix := a.opts.Kind.repoNamePrefix()

	if len(a.opts.RepoName) == 0 {
		a.opts.RepoName = prefix + a.opts.Name
	}

	return a.ask(
		&a.opts.RepoName,
		&survey.Input{
			Message: "GitHub repository name:",
			Help:    "The name of the GitHub repository. The name must start with the " + prefix + " prefix.",
			Default: a.opts.RepoName,
		},
		survey.Required,
		a.expect(
			"startswith="+prefix,
			"The name must start with "+ansi.Color(prefix, "+b"),
		),
	)
}

func (a *asker) askRepoProtocol() error {
	if a.opts.NoGitOrigin || !a.opts.UseGitHub {
		return nil
	}

	if len(a.opts.RepoProtocol) == 0 {
		a.opts.RepoProtocol = "ssh"
	}

	//nolint:lll
	prompt := &survey.Select{
		Message: "Choose git origin protocol:",
		Options: []string{"https", "ssh"},
		Help:    "The protocol used to access the origin git repository can be HTTP or SSH. If you don't have an SSH client, choose HTTPS.",
		Default: a.opts.RepoProtocol,
	}

	return a.ask(&a.opts.RepoProtocol, prompt)
}

func (a *asker) askGoModule() error {
	a.opts.guessGoModule()

	return a.ask(
		&a.opts.GoModule,
		&survey.Input{
			Message: "go module path:",
			Help:    "The go module path to use for the `go mod init` command.",
			Default: a.opts.GoModule,
		},
		survey.Required,
	)
}

func (a *asker) askGoPackage() error {
	a.opts.guessGoPackage()

	return a.ask(
		&a.opts.GoPackage,
		&survey.Input{
			Message: "go package name:",
			Help:    "The go package name used in the generated go source code.",
			Default: a.opts.GoPackage,
		},
		survey.Required,
	)
}

func (a *asker) askNoInstall() error {
	if a.opts.installed {
		return nil
	}

	return a.ask(
		&a.opts.noInstall,
		&survey.Confirm{
			Message: "Disable xk6 install:",
			Default: a.opts.noInstall,
			Help:    "xk6 is a k6 extension development tool that is essential for extension development.",
		},
	)
}

func (a *asker) askAll() error {
	header := ansi.ColorFunc("yellow+b")

	section := func(title string, funcs ...func() error) error {
		a.print("\n%s\n", header(title))

		for _, fun := range funcs {
			if err := fun(); err != nil {
				return err
			}

			a.opts.update()
		}

		return nil
	}

	var err error

	err = section("General options",
		a.askKind, a.askName, a.askSummary, a.askDir, a.askNoInstall,
	)
	if err != nil {
		return err
	}

	err = section("Git repository",
		a.askNoGitInit,
		a.askUseGitHub, a.askRepoOwner, a.askRepoName,
		a.askNoGitOrigin,
		a.askRepoProtocol,
		a.askGitOrigin,
	)
	if err != nil {
		return err
	}

	err = section("Go",
		a.askGoModule,
		a.askGoPackage,
	)
	if err != nil {
		return err
	}

	a.opts.update()

	return nil
}

func (a *asker) askLoop() (bool, error) {
	if a.opts.NoAsk {
		return true, nil
	}

	var err error

	for ok := false; !ok && err == nil; ok, err = a.askConfirm() {
		a.print("%s\n", ansi.Color("Create k6 extension", "yellow+b"))

		if err = a.askAll(); err != nil {
			break
		}

		a.print("\n%s\n\n", ansi.Color("✓ Confirmation", "yellow+b"))
	}

	if err != nil {
		if err.Error() == "interrupt" {
			return false, nil
		}

		return false, err
	}

	a.opts.guessPrimaryClass()

	return true, nil
}

func (a *asker) expect(tag, msg string) survey.Validator {
	return func(value interface{}) error {
		err := a.validate.Var(value, tag)
		if err == nil {
			return nil
		}

		var verr validator.ValidationErrors

		if !errors.As(err, &verr) {
			return err
		}

		return errors.New(msg)
	}
}

func (a *asker) ask(
	response interface{},
	prompt survey.Prompt,
	validators ...survey.Validator,
) error {
	return survey.AskOne(
		prompt,
		response,
		survey.WithValidator(survey.ComposeValidators(validators...)),
	)
}
