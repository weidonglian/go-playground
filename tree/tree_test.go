package tree

import (
	"testing"
)

func TestSame(t *testing.T) {
	cases := []struct {
		k1, k2 int
		expect bool
	}{
		{1, 2, false},
		{3, 3, true},
		{0, 0, true},
		{1, 1, true},
		{5, 10, false},
	}

	for _, c := range cases {
		if c.expect != Same(New(c.k1), New(c.k2)) {
			t.Errorf("Same(New(%v), New(%v)) != %v", c.k1, c.k2, c.expect)
		}
	}
}
