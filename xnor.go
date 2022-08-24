package main

import "fmt"

// Activate the all bits after most significant bit
func setAllBit(num int) int {
	var n int = num
	n = n | n >> 1
	n = n | n >> 2
	n = n | n >> 4
	n = n | n >> 8
	n = n | n >> 16
	return n
}
// Perform xnor operation of given two numbers
func xnor(x, y int) {
	var result int = 0
	if x > y {
		// When x is greater than y
		result = (setAllBit(x) ^ x) ^ y
	} else {
		// When y is greater than x
		result = (setAllBit(y) ^ y) ^ x
	}
	// Display result
	res := fmt.Sprintf("%x", result)
	fmt.Println(res)
}