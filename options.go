package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2/core"
	"github.com/iancoleman/strcase"
)

type options struct {
	Dir string `json:"dir,omitempty"`

	Kind         kind   `json:"kind,omitempty"`
	Name         string `json:"name,omitempty"`
	Summary      string `json:"summary,omitempty"`
	GitOrigin    string `json:"gitOrigin,omitempty"`
	UseGitHub    bool   `json:"useGitHub,omitempty"`
	RepoOwner    string `json:"repoOwner,omitempty"`
	RepoName     string `json:"repoName,omitempty"`
	RepoProtocol string `json:"repoProtocol,omitempty"`
	GoModule     string `json:"goModule,omitempty"`
	GoPackage    string `json:"goPackage,omitempty"`

	NoGitInit   bool `json:"noGitInit,omitempty"`
	NoGitOrigin bool `json:"noGitOrigin,omitempty"`
	NoAsk       bool `json:"noAsk,omitempty"`

	PrimaryClass string `json:"PrimaryClass,omitempty"`
	EnvPrefix    string `json:"envPrefix,omitempty"`

	installed bool
	noInstall bool
	debug     bool
}

func (opts *options) guessUseGitHub() {
	opts.UseGitHub = opts.UseGitHub ||
		((len(opts.RepoOwner) != 0) && (len(opts.RepoName) != 0) && (len(opts.RepoProtocol) != 0))
}

func (opts *options) guessKind() {
	if len(opts.Kind) != 0 || len(opts.Dir) == 0 {
		return
	}

	dir := filepath.Base(opts.Dir)

	if strings.HasPrefix(dir, prefixJavaScript) {
		if strings.HasPrefix(dir, prefixOutput) {
			opts.Kind = output
		} else {
			opts.Kind = javascript
		}
	} else {
		opts.Kind = javascript
	}
}

func (opts *options) guessName() {
	if len(opts.Name) != 0 || len(opts.Dir) == 0 {
		return
	}

	dir := filepath.Base(opts.Dir)

	if strings.HasPrefix(dir, "xk6-") {
		if strings.HasPrefix(dir, "xk6-output-") {
			opts.Name = strings.TrimPrefix(dir, "xk6-output-")
		} else {
			opts.Name = strings.TrimPrefix(dir, "xk6-")
		}
	}
}

func (opts *options) guessRepoName() {
	if len(opts.RepoName) != 0 {
		return
	}

	opts.RepoName = opts.Kind.repoNamePrefix() + opts.Name
}

func (opts *options) guessGoModule() {
	if len(opts.GoModule) != 0 {
		return
	}

	opts.guessUseGitHub()
	opts.guessName()

	if len(opts.RepoOwner) != 0 && len(opts.RepoName) != 0 {
		opts.GoModule = path.Join("github.com", opts.RepoOwner, opts.RepoName)
	} else if len(opts.Name) != 0 {
		opts.GoModule = opts.Kind.repoNamePrefix() + opts.Name
	}
}

func (opts *options) guessGoPackage() {
	if len(opts.GoPackage) != 0 || len(opts.Name) == 0 {
		return
	}

	opts.GoPackage = strcase.ToSnake(opts.Name)
}

func (opts *options) guessGitOrigin() {
	opts.guessUseGitHub()

	if len(opts.GitOrigin) != 0 || !opts.UseGitHub {
		return
	}

	if len(opts.RepoOwner) == 0 || len(opts.RepoName) == 0 || len(opts.RepoProtocol) == 0 {
		return
	}

	var format string
	if opts.RepoProtocol == "ssh" {
		format = "git@github.com:%s/%s.git"
	} else {
		format = "https://github.com/%s/%s.git"
	}

	opts.GitOrigin = fmt.Sprintf(format, opts.RepoOwner, opts.RepoName)
}

func (opts *options) guessPrimaryClass() {
	if len(opts.PrimaryClass) != 0 {
		return
	}

	opts.PrimaryClass = strcase.ToCamel(opts.Name)
}

func (opts *options) guessDir() {
	if len(opts.Dir) != 0 {
		return
	}

	opts.Dir = opts.Kind.repoNamePrefix() + opts.Name
}

func (opts *options) update() {
	from := opts.RepoName
	if len(from) == 0 {
		from = opts.Kind.repoNamePrefix() + "_" + opts.Name
	}

	opts.EnvPrefix = strcase.ToScreamingDelimited(from, '_', "xk6", true)
}

func (opts *options) guess() {
	opts.guessKind()
	opts.guessName()
	opts.guessRepoName()
	opts.guessUseGitHub()
	opts.guessGoModule()
	opts.guessGoPackage()
	opts.guessGitOrigin()
	opts.guessPrimaryClass()
	opts.guessDir()
}

func (opts *options) toMap() (map[string]interface{}, error) {
	buff, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}

	err = json.Unmarshal(buff, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

type kind string

func (k kind) repoNamePrefix() string {
	if k == javascript {
		return prefixJavaScript
	}

	return prefixOutput
}

func (k *kind) WriteAnswer(_ string, value interface{}) error {
	if v, ok := value.(core.OptionAnswer); ok {
		*k = (kind)(v.Value)
	} else if v, ok := value.(string); ok {
		*k = (kind)(v)
	} else {
		return errors.ErrUnsupported
	}

	return nil
}

const (
	javascript kind = "JavaScript"
	output     kind = "Output"

	prefixJavaScript = "xk6-"
	prefixOutput     = prefixJavaScript + "output-"
)
