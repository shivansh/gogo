func main() {
	a := [2]int{}
	a[0] = 1
	for {
		if a[0] == 1 {
			break
		}
	}
	printInt a[0]
	return
}
