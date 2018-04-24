func main() {
	str := "Enter number of rows: "
        newlineStr := "\n"
        spaceStr := " "
        var rows int

        printStr str
        scanInt rows
        printStr newlineStr
        for i := rows; i >= 1; i-- {
                for j := 1 ; j <= i; j++ {
                        printInt j
                        printStr spaceStr
                }
                printStr newlineStr
        }
        return
}
