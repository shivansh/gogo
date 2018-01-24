CC=      go
BIN=     ./bin
SRC=     ./src
GOCCDIR= errors lexer parser token util
CLEANDIR=$(addprefix $(SRC)/, $(GOCCDIR))

all:
	make deps
	make lexer
	make codegen

.PHONY: lexer clean

deps:
	gocc -o $(SRC) $(SRC)/lang.bnf

lexer: $(SRC)/lexer.go
	mkdir -p $(BIN)
	$(CC) build -o $(BIN)/$@ $<

codegen: $(SRC)/codegen.go
	mkdir -p $(BIN)
	$(CC) build -o $(BIN)/$@ $<

clean:
	rm -rf $(CLEANDIR)
	rm -rf $(BIN)
