// This file implements the next-use register allocation heuristic.

package tac

const (
	// Variables which are dead have their next-use set to MaxInt.
	MaxInt = int(^uint(0) >> 1)
	MinInt = -MaxInt - 1
)

type UseInfo struct {
	// The Name field is used in two different contexts -
	//	- when dealing with priority queue of registers, it is the name
	//	  of a register
	//	- when dealing with lookup tables, it is the name of a variable
	Name string
	// Nextuse determines the next usage (line number) of a varible. If the
	// variable is dead, its Nextuse is set to MaxInt.
	Nextuse int
}

// Next-use allocation heuristic
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Traverse the statements in a basic block from bottom-up, while updating
// the next-use symbol table using the following algorithm (x = y op z) -
// 	Step 1: attach to i'th line in NextUseTab information currently
// 		in symbol table (nuSymTab) about variables x, y and z.
// 	Step 2: mark x as dead and no next-use in nuSymTab.
// 	Step 3: mark y and z to be live and set their next-use to i in nuSymTab.
func (blk Blk) EvalNextUseInfo() {
	// nuSymTab is a symbol table to track next-use information
	// corresponding to all the variables.
	nuSymTab := make(map[string]int)
	for i := len(blk.Stmts) - 1; i >= 0; i-- {
		switch blk.Stmts[i].Op {
		case "label", "func":
			continue
		}
		s := []string{blk.Stmts[i].Dst}
		for _, v := range blk.Stmts[i].Src {
			switch v := v.(type) {
			case Str:
				s = append(s, v.StrVal())
			}
		}
		// Step 1
		for _, v := range s {
			if _, ok := nuSymTab[v]; !ok {
				nuSymTab[v] = MaxInt
			}
			blk.NextUseTab[i] = append(blk.NextUseTab[i], UseInfo{v, nuSymTab[v]})
		}
		// Step 2
		nuSymTab[s[0]] = MaxInt
		// Step 3
		for _, v := range blk.Stmts[i].Src {
			nuSymTab[v.StrVal()] = i
		}
	}
}

// FindNextUse returns the next-use information corresponding to
// a variable "name" available in line number "line" of the table.
func (blk Blk) FindNextUse(line int, name string) int {
	for _, v := range blk.NextUseTab[line] {
		if v.Name == name {
			return v.Nextuse
		}
	}
	return MaxInt
}
