CC=      go
BIN=     ./bin
SRC=     ./src
GOCCDIR= errors lexer parser token util
CLEANDIR=$(addprefix $(SRC)/, $(GOCCDIR))

all:
	make lexer

.PHONY: lexer clean

lexer: $(SRC)/lexer.go
	mkdir -p $(BIN)
	gocc -o $(SRC) $(SRC)/lang.bnf
	$(CC) build -o $(BIN)/$@ $<

clean:
	rm -rf $(CLEANDIR)
	rm -rf $(BIN)
