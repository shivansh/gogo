func checkprime(i int) (int){
	for j:=2; j<i; j++{
		if (i%j == 0) {
			return 1
		}
	}
	return 0
}

func main() {
	input := "Give input"
	startSen := "The number "
	truePrime := " is prime"
	falsePrime := " is not prime"
	newline := "\n"
	var a int 
	var in int
	printStr input
	scanInt in
	for i:=2; i<in; i++{
		a = checkprime(i)
		printStr startSen
		printInt i
		if (a==1){
			printStr falsePrime
			printStr newline
		} else{
			printStr truePrime
			printStr newline
		}
	}
	return
}