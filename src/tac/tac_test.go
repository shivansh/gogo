package tac

import (
	"log"
	"os"
	"testing"
)

func TestGenTAC(t *testing.T) {
	type args struct {
		irfile string
	}
	tests := []struct {
		name string
		args args
	}{{
		name: "Test1",
		args: args{"../../test/test1.ir"},
	},
	}
	for _, tt := range tests {
		file, err := os.Open(tt.args.irfile)
		if err != nil {
			log.Fatal(err)
		}
		tac := GenTAC(file)
		if tac[0].Stmts[0].Op != "label" {
			t.Errorf("Expected: %s, Got: %s", "label", tac[0].Stmts[0].Op)
		}
	}
}
