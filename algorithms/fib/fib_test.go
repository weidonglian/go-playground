package fib

import (
	"fmt"
	"testing"
)

func TestFib(t *testing.T) {
	evs := [10]uint{0, 1, 1, 2, 3, 5, 8, 13, 21, 34}
	ch := Fib(len(evs))
	i := 0
	for v := range ch {
		if evs[i] != v {
			t.Errorf("Fib(%v) for %d, expected:%v, but %v", len(evs), i, evs[i], v)
		}
		i++
	}
}

func BenchmarkFib10(b *testing.B) {
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		for v := range Fib(10) {
			fmt.Printf("The value is %v", v)
		}
	}
}

func ExampleFib() {
	ch := Fib(10)
	for v := range ch {
		fmt.Printf("%v, ", v)
	}
	// Output: 0, 1, 1, 2, 3, 5, 8, 13, 21, 34,
}
