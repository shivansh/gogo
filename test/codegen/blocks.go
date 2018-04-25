func main() {
	withinBlock := "\nInside the block\nValue of a: "
	outsideBlock := "\nOutside the block\nValue of a:"

	a := 1
	printStr outsideBlock
	printInt a
	{
		a := 4
		printStr withinBlock
		printInt a
	}
	printStr outsideBlock
	printInt a
	
	return
}
