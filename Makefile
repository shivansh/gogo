CC=     go
BIN=    ./bin
SRC=    ./src

all:
	make lexer
	make parser

.PHONY: lexer parser clean

lexer: $(SRC)/lexer.go
	mkdir -p $(BIN)
	$(CC) build -o $(BIN)/$@ $<

parser: $(SRC)/parser.go
	mkdir -p $(BIN)
	$(CC) build -o $(BIN)/$@ $<

clean:
	rm -rf bin
