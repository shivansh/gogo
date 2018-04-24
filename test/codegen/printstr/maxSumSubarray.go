func main() {
	a := [1000]int{}
	var length int
	str1 := "Enter length of the array : "
	printStr str1
	scanInt length
	newline := "\n"
	str2 := "Enter "
	printStr str2
	printInt length
	str3 := " integers :\n"
	printStr str3
	var x int
	for j := 0; j<length; j++ {
		scanInt x
		a[j] = x
	}
	sum := [1000]int{}
	sum[0] = a[0]
	var maxsum int = sum[0]
	for i := 1; i<length; i++ {
		if sum[i-1] < 0 {
			sum[i] = a[i]
		} else {
			sum[i] = sum[i-1] + a[i]
		}
		if maxsum < sum[i] {
			maxsum = sum[i]
		}
	}
	str4 := "Sum of maximum sum subarray = "
	printStr str4
	printInt maxsum
	printStr newline
	return
	// testcase := -2, -3, 4, -1, -2, 1, 5, -3
}