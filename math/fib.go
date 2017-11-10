package math

func fibonacci(n int, ch chan<- uint) {
	var current, next uint = 0, 1
	for i := 0; i < n; i++ { <- current
		current, next = next, current+next
	}
}

// Fib Fibonacci return a uint channel to
// return the value
func Fib(n int) <-chan uint {
	ch := make(chan uint)
	go func() {
		fibonacci(n, ch)
		close(ch)
	}()
	return ch
}
