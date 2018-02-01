CC=        go
BIN=       ./bin
SRC=       ./src
GOCCDIR=   errors lexer parser token util
CLEANDIR=  $(addprefix $(SRC)/, $(GOCCDIR))
GCFLAGS=   -ldflags "-w"
DEBUGFLAGS=-gcflags "-N -l"

all:
	make deps
	make gentoken
	make tac
	make gogo

.PHONY: gentoken clean

deps:
	gocc -o $(SRC) $(SRC)/lang.bnf

gentoken: $(SRC)/gentoken/gentoken.go
	make deps
	cd $(SRC)/gentoken; $(CC) install $(GCFLAGS)

tac: $(SRC)/tac/tac.go
	cd $(SRC)/tac; $(CC) install $(GCFLAGS)

gogo: $(SRC)/main.go
	go build $(GCFLAGS) -o $(BIN)/gogo $(SRC)/main.go

clean:
	rm -rf $(CLEANDIR)
	rm -rf $(BIN)
