package suffixarray

import (
	"fmt"
	"log"
	"sort"

	bigarray "github.com/team-spectre/go-bigarray"
)

type compareResult byte

const (
	_ compareResult = iota
	goLeft
	goRight
	stopHere
)

type searchState struct {
	text   *Text
	sa     *SuffixArray
	lcplr  bigarray.BigArray
	phrase string
	index  uint64
	height uint64
	lo     uint64
	hi     uint64
}

func (state *searchState) compare(where uint64) (compareResult, uint64, error) {
	if where < state.lo || where > state.hi {
		panic("BUG")
	}

	pos, err := state.sa.PositionAt(where)
	if err != nil {
		return 0, 0, fmt.Errorf("[%d/%d] sa.PositionAt %d: %v", state.lo, state.hi, where, err)
	}
	if gDebug {
		log.Printf("debug: [%d/%d] height=%d sa.PositionAt[%d]=%d", state.lo, state.hi, state.height, where, pos)
	}

	i := state.height
	iter := state.text.Iterate(pos+state.height, state.text.Len())
	result := compareResult(0)
	for iter.Next() && i < uint64(len(state.phrase)) {
		ch0 := uint64(state.phrase[i])
		ch1 := iter.Symbol()
		if ch0 < ch1 {
			result = goLeft
			break
		}
		if ch0 > ch1 {
			result = goRight
			break
		}
		i++
	}
	err = iter.Close()
	if err != nil {
		return 0, 0, err
	}

	if result == 0 {
		if i < uint64(len(state.phrase)) {
			result = goRight
		} else {
			result = stopHere
		}
	}
	return result, pos, nil
}

// Search performs a binary search on the suffix array, using the provided
// LCP-LR array to reduce the time requirements to O(m + log n).  Returns the
// list of offsets into the text which begin with the given phrase.
func Search(text *Text, sa *SuffixArray, lcplr bigarray.BigArray, phrase string) ([]uint64, error) {
	state := searchState{
		text:   text,
		sa:     sa,
		lcplr:  lcplr,
		phrase: phrase,
		index:  0,
		lo:     0,
		hi:     sa.Len() - 1,
	}

	var where, pos uint64
	var cmp compareResult
	var err error
	found := false
	for state.hi >= (state.lo+3) && !found {
		where = state.lo + (state.hi-state.lo)/2

		state.height, err = lcplr.ValueAt(state.index)
		if err != nil {
			return nil, err
		}

		cmp, pos, err = state.compare(where)
		if err != nil {
			return nil, err
		}

		switch cmp {
		case goLeft:
			state.index = 2*state.index + 1
			state.hi = where - 1

		case goRight:
			state.index = 2*state.index + 2
			state.lo = where + 1

		case stopHere:
			found = true
		}
	}

	for state.hi >= state.lo && !found {
		where = state.lo

		cmp, pos, err = state.compare(where)
		if err != nil {
			return nil, err
		}

		switch cmp {
		case goLeft:
			state.hi = where - 1

		case goRight:
			state.lo = where + 1

		case stopHere:
			found = true
		}
	}

	if !found {
		return nil, nil
	}

	results := make([]uint64, 0, 16)
	results = append(results, pos)

	lowerBound := where
	for lowerBound > state.lo {
		lowerBound--

		cmp, pos, err = state.compare(lowerBound)
		if err != nil {
			return nil, err
		}

		if cmp != stopHere {
			break
		}

		results = append(results, pos)
	}

	upperBound := where
	for upperBound < state.hi {
		upperBound++

		cmp, pos, err = state.compare(upperBound)
		if err != nil {
			return nil, err
		}

		if cmp != stopHere {
			break
		}

		results = append(results, pos)
	}

	sort.Sort(byU64(results))
	return results, nil
}

type byU64 []uint64

func (x byU64) Len() int           { return len(x) }
func (x byU64) Less(i, j int) bool { return x[i] < x[j] }
func (x byU64) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

var _ sort.Interface = byU64(nil)
