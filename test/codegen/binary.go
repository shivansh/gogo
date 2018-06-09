package main

func main(){
	var a int
	binary := [10]int{}
	startSen := "Give input number less than 1024\n"
	newline := "\n"
	binaryLine := "The binary representation of the given number is \n"


	printStr startSen
	scanInt a
	printStr newline
	printStr binaryLine
	i := 0
	for ; a>0; {
		binary[i] = a%2
		a = a/2
		i = i+1 
	}
	for j:=i-1; j>=0; j-- {
		printInt binary[j]
	}
	return
}
