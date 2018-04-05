func main() {
	a := [2]int{}
	a[0] = 1
	for {
		if a[0] == 2 {
			break
		}
	}
	return
}

// func, main
// decl, a, 2
// from, t0, a, 1
// =, t0, 1
// into, a, a,  1, t0
// label, l3
// from, t1, a, 1
// printInt, t1, t1
// bne, l0, t1, 2
// =, t2, 1
// j, l1
// label, l0
// =, t2, 0
// label, l1
// blt, l2, t2, 1
// j, l4
// label, l2
// j, l3
// label, l4
// ret,
