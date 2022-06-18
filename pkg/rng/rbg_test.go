package rng

import "testing"

func TestRBG_Bags(t *testing.T) {
	one := 1
	two := 2
	three := 3

	bag := NewRBG[int](1337, func() []*int {
		return []*int{&one, &two, &three}
	})

	n := bag.Next()

	if n != &one && n != &two && n != &three {
		t.Fatal("expected bag element to be one of one, two or three")
	}

	n = bag.Next()

	if n != &one && n != &two && n != &three {
		t.Fatal("expected bag element to be one of one, two or three")
	}

	n = bag.Next()

	if n != &one && n != &two && n != &three {
		t.Fatal("expected bag element to be one of one, two or three")
	}

	n = bag.Next()

	if n != &one && n != &two && n != &three {
		t.Fatal("expected bag element to be one of one, two or three")
	}

	n = bag.Next()

	if n != &one && n != &two && n != &three {
		t.Fatal("expected bag element to be one of one, two or three")
	}
}
