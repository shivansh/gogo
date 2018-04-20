func sum(n int) int {
        if n == 0 {
                return 0
        }
        printInt n
        x := sum(n-1)
        return 1 + x
}

func main() {
        x := sum(5)
        return
}
