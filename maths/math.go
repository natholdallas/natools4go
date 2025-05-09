package maths

// use high performance way to divide then ceil
func CeilDivide(a, b int64) int64 {
	if (a > 0 && b > 0) || (a < 0 && b < 0) {
		return (a + b - 1) / b
	}
	return (a + b + 1) / b
}

// make split from numbers
// example: SplitDigits(10), it returns [1,2,3,4,5,6,7,8,9,10]
func SplitDigits(n int64) []int64 {
	if n == 0 {
		return []int64{0}
	}
	if n < 0 {
		n = -n
	}
	count := 0
	for temp := n; temp != 0; temp /= 10 {
		count++
	}
	digits := make([]int64, count)
	for i := count - 1; i >= 0; i-- {
		digits[i] = n % 10
		n /= 10
	}
	return digits
}
