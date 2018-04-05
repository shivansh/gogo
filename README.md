# gogo
Go to MIPS compiler implemented in Go. Made as a course project for CS335 (Compiler Design).

<p align="center">
  <img alt="Logo" src="gopher.svg">
</p>

*Gopher vector imported from [egonelbre/gophers
](https://github.com/egonelbre/gophers).*

- - -

## Components

| Component | Demo | Status |
|:------------------------:|:----------------------------------------------------------------------------------------------------------------------------------:|:------------------:|
| Token generation / Lexer | [`test1.out`](test/lexer/test1.out) | :heavy_check_mark: |
| Parser | [`struct.go`](test/parser/struct.go) :arrow_right: [`struct.html`](https://shivanshrai84.gitlab.io/staticPages/assets/struct.html) | :heavy_check_mark: |
| IR generation | [`scope.go`](test/codegen/scope.go) :arrow_right: [`scope.ir`](test/codegen/scope.ir) | :heavy_check_mark: |
| Code generation | [`pascalTriangle.ir`](test/ir/pascalTriangle.ir) :arrow_right: [`pascalTriangle.asm`](test/ir/pascalTriangle.asm) | :heavy_check_mark: |

The file [main.go](src/main.go) contains routines described as follows corresponding to each component -

|    Routine   | Description                                                                                      |
|:------------:|--------------------------------------------------------------------------------------------------|
| `GenToken()` | Generates the tokens returned by lexer from the input program                                    |
|  `GenAsmFromIR()`  | Generates the assembly code using the IR generated from the input program                        |
|  `GenHTML()` | Generates the rightmost derivations used in the bottom-up parsing and pretty-prints them in HTML |
|  `GenAsm()`  | GenAsm generates the assembly code from the input program                                        |

## Setting up
Run `./scripts/setup.sh` from the root directory of the project to set up the pre-commit git hooks.

## Dependencies
* [gocc](https://github.com/goccmack/gocc)

## Build
The following should generate relevant binaries inside the directory `bin` -
```
make
```

Alternatively, individual components can be built via -
```
make deps
make gentoken
make tac
make gogo
```

The generated binary `bin/gogo` can be used to generate `(Tokens | Assembly | HTML)` from the corresponding `(go | IR | go)` files -
```
bin/gogo test.go
```

## Testing
The [tests](test) can be built via -
```
make test
```
