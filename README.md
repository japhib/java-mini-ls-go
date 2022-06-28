# java-mini-ls-go

A lightweight language server for Java, written in Go. Goal is to make a reasonably full language server implementation that can be run effectively on resource-constrained systems like Docker containers, with fast startup time and as little CPU/memory usage as possible.

Still in very early development/proof-of-concept stage.

In the future I'd like to experiment with making parts of the language server modular & distribute-able. For example, instead of each local instance of the language server having a full in-memory copy of all the data structures representing the Java stdlib, what if that was instead managed by a central stdlib server? This has the potential to even further reduce memory usage and startup time.

## Features implemented

- Document symbols
- Syntax errors

## Features planned for proof-of-concept

- Go to definition (for user-defined stuff only)
- Find references (for user-defined stuff only)
- Automatically parse all files in project on startup
- Errors on undefined variable usage
- Type checking (no generics though)
- Hover support for built-in functions (parse the stdlib docs)
- Fix imports action
- Priority order for certain things, like java.util.List taking precedence over java.awt.List

# Credit

- github.com/antlr/antlr4 for parsing
- go.lsp.dev/protocol for the LSP implementation
- github.com/micnncim/protocol-buffers-language-server for showing me how to use go.lsp.dev/protocol

Java spec: https://docs.oracle.com/javase/specs/jls/se18/html/jls-15.html