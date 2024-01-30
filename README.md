# starlark-lsp

A language server for [Starlark][starlark], a Python-inspired configuration language.

Starlark-lsp uses [go.lsp.dev][] and [Tree sitter][] as its main dependencies to implement the LSP/JSON-RPC protocol and Starlark language analysis, respectively. It can either be used as a standalone executable (see pkg/cli) or as a Go library.

## Installing

Ensure you have Go 1.18 or greater installed, then check out this repository and run `make install`.

## CLI

The main command for starlark-lsp is `starlark-lsp start`:

```
Start the Starlark LSP server.

By default, the server will run in stdio mode: requests should be written to
stdin and responses will be written to stdout. (All logging is _always_ done
to stderr.)

For socket mode, pass the --address option.

Usage:
  starlark-lsp start [flags]

Examples:

# Launch in stdio mode with extra logging
starlark-lsp start --verbose

# Listen on all interfaces on port 8765
starlark-lsp start --address=":8765"

# Provide type-stub style files to parse and treat as additional language
# built-ins. If path is a directory, treat files and directories inside
# like python modules: subdir/__init__.py and subdir.py define a subdir module.
starlark-lsp start --builtin-paths "foo.py" --builtin-paths "/tmp/modules"

Flags:
      --address string              Address (hostname:port) to listen on
      --builtin-paths stringArray   Paths to files and directories to parse and treat as additional language builtins
  -h, --help                        help for start

Global Flags:
      --debug     Enable debug logging
      --verbose   Enable verbose logging
```

## Current Status
Initial version by [Tilt][]. 

Starlark-lsp is bundled and used bywith the `ak lsp` command as part of the [`Autokitteh`` VS Code extension][ext].
<!--
The `Tiltfile` in this repository can be used while developing the language server functionality for the `Tiltfile` extension. For more information on how to contribute to the extension, see the [CONTRIBUTING.md][] file in the [vscode-ak][] repository.
-->

[starlark]: https://docs.bazel.build/versions/main/skylark/language.html
[go.lsp.dev]: https://go.lsp.dev/
[Tree sitter]: https://tree-sitter.github.io/tree-sitter/
[Tilt]: https://tilt.dev/
[Autokitteh]: https://www.autokitteh.com/
[ext]: https://marketplace.visualstudio.com/items?itemName=autokitteh
<!--
[CONTRIBUTING.md]: https://github.com/tilt-dev/vscode-tilt/blob/main/CONTRIBUTING.md#language-server
-->
[vscode-ak]: https://github.com/autokitteh/vscode-extension
