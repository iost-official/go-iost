package host

import (
	"testing"
)

func TestCtx(t *testing.T) {
	c := NewContext(nil)
	c2 := NewContext(c)
	c3 := NewContext(c2)

	c2.Set("a", 1)
	if c3.Value("a") != 1 {
		t.Fatal(c3.Value("a"))
	}
	if c.Value("a") != nil {
		t.Fatal(c.Value("a"))
	}

	c3.GSet("b", 2)
	if c.GValue("b") != 2 {
		t.Fatal(c.GValue("b"))
	}
}
