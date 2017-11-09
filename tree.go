// Exercise: Equivalent Binary Trees see https://tour.golang.org/concurrency/8
// 1. Implement the Walk function.

// 2. Test the Walk function.

// The function tree.New(k) constructs a randomly-structured (but always sorted) binary tree holding the values k, 2k, 3k, ..., 10k.

// Create a new channel ch and kick off the walker:

// go Walk(tree.New(1), ch)
// Then read and print 10 values from the channel. It should be the numbers 1, 2, 3, ..., 10.

// 3. Implement the Same function using Walk to determine whether t1 and t2 store the same values.

// 4. Test the Same function.

// Same(tree.New(1), tree.New(1)) should return true, and Same(tree.New(1), tree.New(2)) should return false.

package main

import (
	"math/rand"
	"fmt"
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
		v1, ok1 := <- ch1
		v2, ok2 := <- ch2
		if !ok1 || !ok2 {
			return ok1 == ok2
		}

		if v1 != v2 {
			break
		}
	}
	return false
}

func main() {
	fmt.Println(Same(New(2), New(2)))
}