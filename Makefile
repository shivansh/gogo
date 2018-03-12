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

deps:
	mkdir -p tmp
	gocc -o tmp $(SRC)/lang.bnf

gentoken: $(SRC)/gentoken/gentoken.go
	make deps
	cd $(SRC)/gentoken; $(CC) install $(GCFLAGS)

tac: $(SRC)/tac/tac.go
	cd $(SRC)/tac; $(CC) install $(GCFLAGS)

gogo: $(SRC)/main.go
	go build $(GCFLAGS) -o $(BIN)/gogo $(SRC)/main.go

parser: $(SRC)/parser/read_parser.go $(SRC)/parser/test_parser.go
	go run $(SRC)/parser/productions.go | tac > $(SRC)/parser/output.txt
	go run $(SRC)/parser/gen_html.go > $(SRC)/parser/output.html

test:
	scripts/run-tests.sh

clean:
	rm -rf $(CLEANDIR)
	rm -rf $(BIN)
