#!/usr/bin/env bash
java -jar ~/projects/antlr-4.10.1-complete.jar -package javaparser -Dlanguage=Go -o ../javaparser JavaParser.g4 JavaLexer.g4