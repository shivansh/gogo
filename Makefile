CC=        go
BIN=       ./bin
SRC=       ./src
CLEANDIR=  $(SRC)/goccgen
GCFLAGS=   -ldflags "-w"
DEBUGFLAGS=-gcflags "-N -l"

all:
	make deps
	make gogo

.PHONY: scanner tac gogo clean test

deps: $(SRC)/lang.bnf
	scripts/check-gocc.sh

scanner: $(SRC)/scanner/gentoken.go
	cd $(SRC)/scanner; $(CC) install $(GCFLAGS)

tac: $(SRC)/tac/tac.go
	cd $(SRC)/tac; $(CC) install $(GCFLAGS)

gogo: $(SRC)/main.go
	mkdir -p tmp
	$(CC) build $(GCFLAGS) -o $(BIN)/gogo $(SRC)/main.go
	@echo --- [ Regenerating runtime ] ------------------------------
	$(BIN)/gogo -runtime

test:
	scripts/run-tests.sh -r2s
	scripts/run-tests.sh -r
	make testdiff

testdiff:
	git diff --exit-code ./test

gofmt:
	gofmt -l -s -w ./src

govet:
	go tool vet -methods=false ./src

errcheck:
	go get github.com/kisielk/errcheck
	errcheck -exclude .errcheck-ignore ./src/...

travis:
	make
	make test
	make govet
	make gofmt
	make errcheck

clean:
	rm -rf $(CLEANDIR)
	rm -rf $(BIN)
