// Exercise: Equivalent Binary Trees see https://tour.golang.org/concurrency/8
package tree

import (
	"math/rand"
)

type Tree struct {
	Left  *Tree
	Value int
	Right *Tree
}

func New(k int) *Tree {
	var t *Tree
	for _, v := range rand.Perm(10) {
		t = insert(t, v*k)
	}
	return t
}

func insert(t *Tree, v int) *Tree {
	if t == nil {
		return &Tree{nil, v, nil}
	}

	if v <= t.Value {
		t.Left = insert(t.Left, v)
	} else {
		t.Right = insert(t.Right, v)
	}

	return t
}

// Walk walks the tree t sending all values
// from the tree to the channel ch.
func Walk(t *Tree, ch chan int) {
	if t == nil {
		return
	}

	if t.Left != nil {
		Walk(t.Left, ch)
	}

	ch <- t.Value

	if t.Right != nil {
		Walk(t.Right, ch)
	}
}

func Walker(t *Tree) <-chan int {
	ch := make(chan int)
	go func() {
		Walk(t, ch)
		close(ch)
	}()
	return ch
}

// Same determines whether the trees
// t1 and t2 contain the same values.
func Same(t1, t2 *Tree) bool {
	ch1, ch2 := Walker(t1), Walker(t2)
	for {
		v1, ok1 := <-ch1
		v2, ok2 := <-ch2
		if !ok1 || !ok2 {
			return ok1 == ok2
		}

		if v1 != v2 {
			break
		}
	}
	return false
}
