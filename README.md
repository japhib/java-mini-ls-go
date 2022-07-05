# java-mini-ls-go

A lightweight language server for Java, written in Go. Goal is to make a reasonably full language server 
implementation that can be run effectively on resource-constrained systems like Docker containers, with 
fast startup time and as little CPU/memory usage as possible.

Still in very early development/proof-of-concept stage.

In the future I'd like to experiment with making parts of the language server modular & distribute-able. 
For example, instead of each local instance of the language server having a full in-memory copy of all 
the data structures representing the Java stdlib, what if that was instead managed by a central stdlib 
server? This has the potential to even further reduce memory usage and startup time.

## Features implemented

- Document symbols
- Syntax errors
- Shows local variable type on hover

## Features planned for proof-of-concept

- Type checking (no generics though)
- Handle array/varargs parameters
- Go to definition (for user-defined stuff only)
- Find references (for user-defined stuff only)
- Automatically parse all files in project on startup
- Show docs on hover for stdlib classes/methods
- Show signature on hover for user-defined classes/methods

## Limitations

- Ignores generics
- Ignores dependencies
- Definitions/usages gets confused by method/constructor overloads
- Only loads [java.base](https://docs.oracle.com/en/java/javase/17/docs/api/java.base/module-summary.html) module of 
Java standard library (packages like java.lang, java.util)
- Doesn't do fully-qualified class name resolution. So e.g. if you define a class
called `Object`, you'll get all sorts of weird behavior as the language server will
mix up your `Object` class with `java.lang.Object`
- Low unit test coverage (currently 50% or less)

## Ideas for useful features beyond compile errors

It's great when the IDE can show how to 

- Corrections for common mistakes such as:
  - using == on a String
  - Unused local variable
  - Local variable shadows an instance variable
- Common actions such as:
  - Finding and adding imports to top of file
  - Formatting the file (should be configurable)
  - Adding an optional parameter to a method
- Suggestions for when a method can be made static
- Automatic refactors:
  - Symbol renames
  - Extract method
  - Change the signature of a method
  - Change imperative for-loops into FP-style streams
  - Turn assigning/returning if statements into ?: conditional expressions
  - Turn index-based for loops into for-range loops
  - Use new switch expressions & pattern-matching where possible

# Credit

- [Replit](https://replit.com) for the motivation
- [github.com/antlr/antlr4](https://github.com/antlr/antlr4) for parsing
- [go.lsp.dev/protocol](https://github.com/go-language-server/protocol) for the LSP implementation
- [github.com/micnncim/protocol-buffers-language-server](https://github.com/micnncim/protocol-buffers-language-server) for showing me how to use `go.lsp.dev/protocol`
- [Java spec](https://docs.oracle.com/javase/specs/jls/se18/html/jls-15.html)