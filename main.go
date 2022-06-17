package main

import (
	"java-mini-ls-go/parse_utils"
)

func main() {
	parse_utils.Lex("class MyClass {}")

	parse_utils.Parse("class MyClass {}")
}
