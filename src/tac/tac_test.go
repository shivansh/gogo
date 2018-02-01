package tac

import "testing"

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
		tac := GenTAC(tt.args.irfile)
		if tac.stmts[0].op != "func" {
			t.Errorf("Expected: %s, Got: %s", "ret", tac.stmts[9].op)
		}
	}
}
