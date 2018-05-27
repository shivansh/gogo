# gogo
[![Build Status](https://travis-ci.org/shivansh/gogo.svg?branch=master)](https://travis-ci.org/shivansh/gogo)

Go to MIPS compiler implemented in Go. Made as a course project for CS335 (Compiler Design).

<p align="center">
  <img alt="Logo" src="gopher.svg">
</p>

*Gopher vector imported from [egonelbre/gophers
](https://github.com/egonelbre/gophers).*

- - -

## Components

| Component | Demo |
|:------------------------:|:----------------------------------------------------------------------------------------------------------------------------------:|
| Token generation / Lexer | [`test1.out`](test/lexer/test1.out) |
| Parser | [`struct.go`](test/parser/struct.go) :arrow_right: [`struct.html`](https://shivanshrai84.gitlab.io/staticPages/assets/struct.html) |
| IR generation | [`scope.go`](test/codegen/scope.go) :arrow_right: [`scope.ir`](test/codegen/scope.ir) |
| Code generation | [`pascalTriangle.ir`](test/ir/pascalTriangle.ir) :arrow_right: [`pascalTriangle.asm`](test/ir/pascalTriangle.asm) |

## Setting up
Run `./scripts/setup.sh` from the root directory of the project to set up the pre-commit git hooks.

## Dependencies
* [gocc](https://github.com/goccmack/gocc)

## Build
The following should generate relevant binaries inside the directory `bin` -
```
make
```

The generated binary `bin/gogo` can be used as follows -
```
Usage: gogo (-r | -r2s | -s) <filename>
  -p	Generates rightmost derivations used in bottom-up parsing
  -r	Generates IR instructions from go program
  -r2s  Generates the MIPS assembly from IR
  -s	Generates MIPS assembly from go program
```

**NOTE:** The generated MIPS assembly has been tested to work on [SPIM](http://spimsimulator.sourceforge.net/) MIPS32 simulator.

## Testing
The [tests](test) can be built via -
```
make test
```
