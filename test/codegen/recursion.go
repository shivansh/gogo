package main

// PrintNatNums prints the first n natural numbers in decreasing order.
func PrintNatNums(n int) int {
        newline := "\n"

        if n == 0 {
                return 0
        }
        printInt n
        printStr newline
        x := PrintNatNums(n-1)

        return 1 + x
}

func main() {
        PrintNatNums(5)
        return
}
