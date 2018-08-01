package bystander

import (
	"testing"
)

// TestForeachIterator tests the iterator
func TestForeachIterator(t *testing.T) {

	foreach := map[string]string{
		"a": "aa",
		"b": "bb",
		"c": "cc",
		"d": "dd",
	}

	aa := staticForeach([]string{"a1", "a2"})
	bb := staticForeach([]string{"b1", "b2", "b3"})
	cc := staticForeach([]string{"c1"})
	dd := staticForeach([]string{"d1", "d2"})
	vars := map[string]foreachConfig{
		"a": &aa,
		"b": &bb,
		"c": &cc,
		"d": &dd,
	}

	numSeen := 0
	itr := newForeachIter(foreach, vars)
	for itr.Next() {
		m := itr.Value()
		t.Logf("got %v\n", m)
		numSeen++
	}

	// TODO improve testing
	numExpected := 12
	if numSeen != numExpected {
		t.Errorf("want %v; got %v", numExpected, numSeen)
	}
}
