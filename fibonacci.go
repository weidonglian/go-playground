package main

import (
	"fmt"
)

func fibonacci(n int, ch chan<- uint) {
	var current, next uint = 0, 1
	for i := 0; i < n; i++ {
		ch <- current
		current, next = next, current+next
	}
}

// Fib Fibonacci return a uint channel to
// return the value
func fib(n int) <-chan uint {
	ch := make(chan uint)
	go func() {
		fibonacci(n, ch)
		close(ch)
	}()
	return ch
}

func main() {
	ch := fib(10)
	for v := range ch {
		fmt.Println(v)
	}
}
