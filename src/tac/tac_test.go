package tac

import (
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
		args: args{"../../test/ir/logarithm.ir"},
	},
	}
	for _, tt := range tests {
		tac := GenTAC(tt.args.irfile)
		if tac[0].Stmts[0].Op != "#" {
			t.Errorf("Expected: %s, Got: %s", "label", tac[0].Stmts[0].Op)
		}
	}
}
