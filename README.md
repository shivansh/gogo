# gogo
Golang to MIPS compiler made as a course project for CS335 (Compiler Design).

<p align="center">
  <img alt="Logo" src="gopher.svg">
</p>

*Gopher vector imported from [egonelbre/gophers
](https://github.com/egonelbre/gophers).*

- - -

## Setting up
Run `./scripts/setup.sh` from the root directory of the project to set up the pre-commit git hooks.

## Build
The following should generate relevant binaries inside the directory `bin` -
```
make
```

Alternatively, individual components can be built via -
```
make lexer
```

The tokens and their corresponding lexemes can be produced via -
```
./bin/lexer test/calcfile  # Currently based on example calc BNF
```
