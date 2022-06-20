# java-mini-ls-go

A lightweight language server for Java, written in Go. Goal is to make a reasonably full language server implementation that can be run effectively on resource-constrained systems like Docker containers, with fast startup time and as little CPU/memory usage as possible.

In the future I'd like to experiment with making parts of the language server modular & distribute-able. For example, instead of each local instance of the language server having a full in-memory copy of all the data structures representing the Java stdlib, what if that was instead managed by a central stdlib server? This has the potential to even further reduce memory usage and startup time.

# Credit

- github.com/antlr/antlr4 for parsing
- go.lsp.dev/protocol for the LSP implementation
- github.com/micnncim/protocol-buffers-language-server for showing me how to use go.lsp.dev/protocol
- github.com/IWANABETHATGUY/tower-lsp-boilerplate for showing me how to set up an VSCode LSP extension for an arbitrary executable (similar in Rust and Go)
