//nolint:forbidigo
package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/mgutz/ansi"

	"github.com/valyala/fasttemplate"
)

func create(opts *options, stdio *terminal.Stdio) error {
	c, err := newCreator(opts, stdio)
	if err != nil {
		return err
	}

	return c.create()
}

type creator struct {
	*terminal.Stdio
	opts    *options
	spinner *spinner.Spinner
	data    map[string]interface{}

	srcDir string
	debug  []byte
}

func newCreator(opts *options, stdio *terminal.Stdio) (*creator, error) {
	c := new(creator)

	c.Stdio = stdio
	c.opts = opts
	c.spinner = spinner.New(
		spinner.CharSets[28],
		200*time.Millisecond,
		spinner.WithColor("yellow"), spinner.WithWriterFile(os.Stdout),
	)

	var err error
	c.data, err = opts.toMap()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *creator) print(format string, a ...any) {
	fmt.Fprintf(c.Out, format, a...)
}

func (c *creator) run(name string, args ...string) error {
	cmd := exec.Command(name, args...)

	var err error

	c.debug, err = cmd.CombinedOutput()

	return err
}

func (c *creator) runIn(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	var err error

	c.debug, err = cmd.CombinedOutput()

	return err
}

func (c *creator) output(dir string, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)

	if len(dir) != 0 {
		cmd.Dir = dir
	}

	return cmd.Output()
}

func (c *creator) step(msg string, fn func() error) error {
	c.spinner.Suffix = " " + msg
	c.spinner.FinalMSG = ansi.Color("✓", "green") + " " + msg + "\n"
	c.spinner.Start()

	var err error

	defer func() {
		if err != nil {
			c.spinner.FinalMSG = ansi.Color("✗", "red") + " " + msg + "\n"
		}

		c.spinner.Stop()

		if len(c.debug) != 0 && (err != nil || c.opts.debug) {
			c.print(color.New(color.Italic).Sprint(string(c.debug)))
		}

		c.debug = nil
	}()

	err = fn()

	return err
}

func (c *creator) downloadTemplate() error {
	var err error
	var dir string

	suffix := strings.ToLower((string)(c.opts.Kind))

	dir, err = os.MkdirTemp("", "template-"+suffix+"-") //nolint:forbidigo
	if err != nil {
		return err
	}

	repo := fmt.Sprintf("https://github.com/szkiba/xk6-template-%s.git", suffix)

	err = c.run("git", "clone", "--depth", "1", repo, dir)
	if err != nil {
		return err
	}

	err = os.RemoveAll(filepath.Join(dir, ".git")) //nolint:forbidigo
	if err != nil {
		return err
	}

	c.srcDir = dir

	return nil
}

func (c *creator) expandTemplate() error {
	bin, oerr := c.output(c.srcDir, "go", "list", "-m")
	if oerr != nil {
		return oerr
	}

	lines := strings.SplitN(string(bin), "\n", 2)

	srcMod := lines[0]

	oerr = filepath.WalkDir(c.srcDir, func(path string, entry fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}

		if path == c.srcDir {
			return os.Mkdir(c.opts.Dir, 0o750)
		}

		var relSrc string
		var err error

		if relSrc, err = filepath.Rel(c.srcDir, path); err != nil {
			return err
		}

		var buff bytes.Buffer

		if _, err = fasttemplate.ExecuteStd(relSrc, "ˮ", "ˮ", &buff, c.data); err != nil {
			return err
		}

		dst := filepath.Join(c.opts.Dir, buff.String())

		if entry.IsDir() {
			return os.Mkdir(dst, 0o750)
		}

		var bin []byte

		if bin, err = os.ReadFile(filepath.Clean(path)); err != nil {
			return err
		}

		buff.Reset()

		if _, err = fasttemplate.ExecuteStd(string(bin), "ˮ", "ˮ", &buff, c.data); err != nil {
			return err
		}

		content := bytes.ReplaceAll(buff.Bytes(), []byte(srcMod), []byte(c.opts.GoModule))

		return os.WriteFile(dst, content, 0o600)
	})
	if oerr != nil {
		return oerr
	}

	return os.RemoveAll(c.srcDir)
}

func (c *creator) createGitRepository() error {
	if err := c.run("git", "init", c.opts.Dir); err != nil {
		return err
	}

	if !c.opts.NoGitOrigin {
		if err := c.runIn(c.opts.Dir, "git", "remote", "add", "origin", c.opts.GitOrigin); err != nil {
			return err
		}
	}

	return nil
}

func (c *creator) runGoGenerate() error {
	return c.runIn(c.opts.Dir, "go", "generate", "./...")
}

func (c *creator) install() error {
	return c.run("go", "install", "go.k6.io/xk6/cmd/xk6@latest")
}

func (c *creator) commitGitRepository() error {
	if err := c.runIn(c.opts.Dir, "git", "add", "."); err != nil {
		return err
	}

	return c.runIn(c.opts.Dir, "git", "commit", "-m", "Initial commit")
}

func (c *creator) create() error {
	c.print("\n\n%s\n", ansi.Color("Creating extension", "yellow+b"))

	if err := c.step("Download template", c.downloadTemplate); err != nil {
		return err
	}

	if err := c.step("Expand template", c.expandTemplate); err != nil {
		return err
	}

	if !c.opts.NoGitInit {
		if err := c.step("Create git repository", c.createGitRepository); err != nil {
			return err
		}
	}

	if err := c.step("Generate sources", c.runGoGenerate); err != nil {
		return err
	}

	if !c.opts.NoGitInit {
		if err := c.step("Commit git repository", c.commitGitRepository); err != nil {
			return err
		}
	}

	if !c.opts.installed {
		if err := c.step("Install xk6", c.install); err != nil {
			return err
		}
	}

	c.print("\n%s\n",
		ansi.Color("Congratulations, the extension is ready!", "green"),
	)
	c.print("You can find the initial version of your new extension in:\n  %s\n",
		ansi.Color(c.opts.Dir, "yellow"),
	)
	c.print("For more information on extension development, visit:\n  %s\n",
		ansi.Color("https://grafana.com/docs/k6/latest/extensions/create/", "cyan"),
	)

	if c.opts.Kind != javascript {
		return nil
	}

	c.print("Use the following commands to build k6 with the %s extension:\n  %s\n  %s\n",
		ansi.Color(c.opts.Name, "yellow"),
		ansi.Color("cd "+c.opts.Dir, "yellow"),
		ansi.Color(fmt.Sprintf("xk6 build --with %s=.", c.opts.GoModule), "yellow"),
	)

	c.print("You can test the extension with the following command:\n  %s\n",
		ansi.Color("./k6 run test.js", "yellow"),
	)

	c.print(
		`The TypeScript definition of the extension API can be found in:
  %s
The source code and README.md can be regenerated with the following command:
  %s
`,
		ansi.Color("index.d.ts", "yellow"),
		ansi.Color("go generate", "yellow"),
	)
	c.print("See the documentation for more information:\n  %s\n",
		ansi.Color("https://github.com/szkiba/create-k6-extension/", "cyan"),
	)

	return nil
}
