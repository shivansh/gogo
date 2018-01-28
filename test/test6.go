package main

import "fmt"

func max(x, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}

func main() {
	var n int
	fmt.Scanf("%d", &n)
	var num []int
	num = make([]int, n, n)
	for i := 0; i < n; i++ {
		fmt.Scanf("%d", &num[i])
	}
	var max_sum []int
	max_sum = make([]int, n, n)
	max_sum[0] = num[0]
	var maxSum = num[0]
	for i := 1; i < n; i++ {
		if max_sum[i-1] < 0 {
			max_sum[i] = num[i]
		} else {
			max_sum[i] = max_sum[i-1] + num[i]
		}
		maxSum = max(maxSum, max_sum[i])
	}
	fmt.Println(maxSum)
}
