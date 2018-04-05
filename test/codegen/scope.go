func main() {
	arr := [2]int{}
	arr[0] = 4
	{
		arr := [2]int{}
		arr[0] = 5
	}

	a := 1
	{
		a := 4
		a = 2
	}
	a = 4
}
