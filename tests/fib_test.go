package tests

import (
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
