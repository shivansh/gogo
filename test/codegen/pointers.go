func main() {
        x := 1
        y := &x
        z := *y
        // should print 1
        printInt z
        *y = 4
        // should print 4
        printInt x
        return
}
