# create-k6-extension

**Set up a k6 extension development by running one command.**

`create-k6-extension` allows you to create an initial version of your k6 extension by running a single command. After answering a few questions, it will generate an initial working k6 extension. If you prefer command line flags, you can also use `create-k6-extension` in non-interactive mode.

```bash
go run github.com/szkiba/create-k6-extension@latest
```

That's all. After the command runs successfully, a fully prepared working directory is created.

![readme.svg](readme.svg)

## Prerequisites

To use `create-k6-extension`, you need to install [go toolchain](https://go.dev/doc/install) and [git CLI](https://git-scm.com/downloads). You probably already have these, but if you don't, install them.

To develop the extension, you will need the [xk6](https://github.com/grafana/xk6) tool. If you haven't installed it yet, `create-k6-extension` will ask for it and install it automatically if you want.

## Install

It is recommended to use `create-k6-extension` without installation, using the following command:

```bash
go run github.com/szkiba/create-k6-extension@latest
```

If you want to install it, that is also possible. Precompiled binaries can be downloaded and installed from the [Releases](https://github.com/szkiba/create-k6-extension/releases) page.

Installation can also be done with the following command:

```bash
go install github.com/szkiba/create-k6-extension@latest
```

## Usage

`create-k6-extension` can be started with the following command:

```bash
go run github.com/szkiba/create-k6-extension@latest [flags] [directory]
```

or if you installed

```bash
create-k6-extension [flags] [directory]
```

After a successful run, [xk6](https://github.com/grafana/xk6) can be used to build k6 with the extension. `create-k6-extension` will print the xk6 command parameters to use for the build. In the case of a JavaScript extension, you also get a `test.js` file, which can be used to test the extension.

**non-interactive mode**

Use the `--no-ask` flag to activate non-interactive mode. In this case, the answers to the questions can be given using flags. If a mandatory answer is missing, you will receive an error message with the name of the missing flag.

Flags can also be used in interactive mode, then you can set default answers with them.

```
Flags:
      --debug                  enable debug output
      --git-origin string      git origin URL
      --go-module string       go module path
      --go-package string      go package name (default: extension name)
  -h, --help                   print this help message
      --name string            extension name
      --no-ask                 disable interactive questions
      --no-git-init            disable git module initialization
      --no-git-origin          disable setting git origin
      --repo-name string       GitHub repository name
      --repo-owner string      GitHub repository owner
      --repo-protocol string   git repository origin protocol (ssh or https) (default "ssh")
      --summary string         a brief summary of the extension
      --type string            extension type (JavaScript or Output) (default "JavaScript")
      --version                print version
```

## Development

In the case of a JavaScript extension, the API of the extension is contained in the `index.d.ts` file. After modification, go interfaces can be generated from it using the `go generate` command. The extension is developed by implementing these interfaces.

The `go generate` command also updates the documentation in the `README.md` file based on `index.d.ts`. That is, the primary source of the documentation and the API is the `index.d.ts` file.

For more information on the *API-First approach k6 extension development*, see https://github.com/szkiba/tygor

In the case of an Output extension, the development process is simpler, there is no need for a code generation phase. You simply need to implement the functionality of the extension in the `flush` method.

## How It Works

The extension is created based on the template repository corresponding to the type of extension (JavaScript, Output).

The template repositories are downloaded at runtime (by running the `git clone` command).

The JavaScript template repository is https://github.com/szkiba/xk6-template-javascript and the Output template repository is https://github.com/szkiba/xk6-template-output

Templates are simple variable substitution-based template files. Variable substitution is also done in file and directory names.

The character used as a delimiter character for template variable substitution (`ˮ`) is considered a letter, therefore template variable references are also valid identifiers in different programming languages (for example, `ˮnameˮ` is a valid identifier in go and JavaScript). In this way, template repositories are also working k6 extensions, which makes it easy to maintain templates.

It was a design consideration that the development of the created extension does not require the installation of an external tool (except for [xk6](https://github.com/grafana/xk6), which can be installed automatically by `create-k6-extension`).