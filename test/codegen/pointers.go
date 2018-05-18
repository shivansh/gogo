func main() {
        a := 1
        b := 2
        x := &a
        y := *x
        // should print 1
        printInt y
        *x = 4
        // should print 4
        printInt a
        z := &b
        *z = *x
        // should print 4
        printInt b
        return
}
