package demo

import (
	"testing"
)

func TestValue(t *testing.T) {
	n := 4
	s := n / 2
	t.Logf("type: %T, value: %v", n, n)
	t.Logf("type: %T, value: %v", s, s)
}
