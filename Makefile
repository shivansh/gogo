CC=        go
BIN=       ./bin
SRC=       ./src
CLEANDIR=  ./tmp
GCFLAGS=   -ldflags "-w"
DEBUGFLAGS=-gcflags "-N -l"

all:
	make deps
	make gentoken
	make tac
	make gogo

.PHONY: gentoken tac gogo clean test

deps: $(SRC)/lang.bnf
	mkdir -p tmp
	gocc -a -zip -o tmp $(SRC)/lang.bnf

gentoken: $(SRC)/gentoken/gentoken.go
	cd $(SRC)/gentoken; $(CC) install $(GCFLAGS)

tac: $(SRC)/tac/tac.go
	cd $(SRC)/tac; $(CC) install $(GCFLAGS)

gogo: $(SRC)/main.go
	$(CC) build $(GCFLAGS) -o $(BIN)/gogo $(SRC)/main.go

test:
	scripts/run-tests.sh

clean:
	rm -rf $(CLEANDIR)
	rm -rf $(BIN)
