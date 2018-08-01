package bystander

import "fmt"

type foreachIter struct {
	varnames     []string
	varvalues    [][]string
	indices      []int
	currentValue map[string]string
	lastValue    bool
}

func newForeachIter(foreach map[string]string, vars map[string]foreachConfig) *foreachIter {
	varnames := []string{}
	varvalues := [][]string{}
	indices := []int{}
	for k, v := range foreach {
		varnames = append(varnames, v)
		if _, ok := vars[k]; !ok {
			panic(fmt.Sprintf("foreach variable %q does not exist in global vars section", k))
		}
		varvalues = append(varvalues, vars[k].values())
		indices = append(indices, 0)
	}

	s := &foreachIter{
		varnames:  varnames,
		varvalues: varvalues,
		indices:   indices,
	}

	return s
}

func (s *foreachIter) nextHelper(m map[string]string, offset int) bool {
	if offset >= len(s.indices) {
		return true
	}
	i := s.indices[offset]
	k := s.varnames[offset]
	vals := s.varvalues[offset]
	v := vals[i]
	m[k] = v
	rolledOver := s.nextHelper(m, offset+1)
	if rolledOver {
		i++
		if i >= len(vals) {
			i = 0
		} else {
			rolledOver = false
		}
		s.indices[offset] = i
	}
	return rolledOver

}

func (s *foreachIter) Value() map[string]string {
	if s.currentValue == nil {
		panic("Value() called before Next()")
	}
	return s.currentValue
}

func (s *foreachIter) Next() bool {
	if s.lastValue {
		return false
	}
	m := map[string]string{}
	s.lastValue = s.nextHelper(m, 0)
	s.currentValue = m
	return true
}
