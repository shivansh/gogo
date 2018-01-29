CC=        go
BIN=       ./bin
SRC=       ./src
GOCCDIR=   errors lexer parser token util
CLEANDIR=  $(addprefix $(SRC)/, $(GOCCDIR))
GCFLAGS=   -ldflags "-w"
DEBUGFLAGS=-gcflags "-N -l"

all:
	make deps
	make lexer
	make tac

.PHONY: lexer clean

deps:
	gocc -o $(SRC) $(SRC)/lang.bnf

lexer: $(SRC)/lexer.go
	mkdir -p $(BIN)
	$(CC) build $(GCFLAGS) -o $(BIN)/$@ $<

tac: $(SRC)/tac.go
	mkdir -p $(BIN)
	$(CC) build $(GCFLAGS) -o $(BIN)/$@ $<

clean:
	rm -rf $(CLEANDIR)
	rm -rf $(BIN)
