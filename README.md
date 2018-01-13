# gogo
Golang to MIPS compiler made as a course project for CS335 (Compiler Design).

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
make parser
```
