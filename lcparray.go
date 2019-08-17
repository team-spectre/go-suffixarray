package suffixarray

import (
	"bytes"
	"fmt"
	"math"

	bigarray "github.com/team-spectre/go-bigarray"
)

// LCPArray records the Longest Common Prefix between two adjacent suffixes in
// a text's suffix array.
//
// Index n≥1 is lcp(SA[n-1], SA[n]); index 0 is undefined.
//
type LCPArray struct {
	ba bigarray.BigArray
}

// LCPIterator iterates over an LCP array.
type LCPIterator struct {
	impl bigarray.Iterator
}

// NewLCPArray constructs an LCP array.
func NewLCPArray(opts ...Option) (*LCPArray, error) {
	opts = extendOptions(
		opts,
		BytesPerValue(8))

	ba, err := makeBigArray(opts)
	if err != nil {
		return nil, err
	}
	return &LCPArray{ba}, nil
}

// MaxValue returns the maximum value that the LCP array can hold.
func (lcp *LCPArray) MaxValue() uint64 { return lcp.ba.MaxValue() }

// Len returns the length of the LCP array, which is equal to the length of the
// suffix array.
func (lcp *LCPArray) Len() uint64 { return lcp.ba.Len() }

// HeightAt returns lcp(SA[index-1], SA[index]), and is undefined for index 0.
func (lcp *LCPArray) HeightAt(index uint64) (height uint64, err error) {
	return lcp.ba.ValueAt(index)
}

// SetHeightAt assigns the given value to the given index.
func (lcp *LCPArray) SetHeightAt(index uint64, height uint64) error {
	return lcp.ba.SetValueAt(index, height)
}

// Iterate constructs an iterator.
func (lcp *LCPArray) Iterate(i, j uint64) *LCPIterator {
	return &LCPIterator{impl: lcp.ba.Iterate(i, j)}
}

// CopyFrom copies all values from a source LCP array to this array. The two
// arrays must have the same Len().
func (lcp *LCPArray) CopyFrom(src *LCPArray) error {
	return lcp.ba.CopyFrom(src.ba)
}

// Truncate trims the LCPArray to the given length.
func (lcp *LCPArray) Truncate(length uint64) error { return lcp.ba.Truncate(length) }

// Freeze makes the LCPArray read-only.
func (lcp *LCPArray) Freeze() error { return lcp.ba.Freeze() }

// Flush ensures that all pending writes have reached the OS.
func (lcp *LCPArray) Flush() error { return lcp.ba.Flush() }

// Close flushes any writes and frees the resources used by the LCPArray.
func (lcp *LCPArray) Close() error { return lcp.ba.Close() }

// Debug returns a human-friendly debugging representation of the LCPArray.
func (lcp *LCPArray) Debug() string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	buf.WriteByte('.')
	err := bigarray.ForEach(lcp.ba, func(index uint64, height uint64) error {
		if index == 0 {
			return nil
		}
		buf.WriteByte(' ')
		fmt.Fprintf(&buf, "%d", height)
		return nil
	})
	if err != nil {
		panic(err)
	}
	buf.WriteByte(']')
	return buf.String()
}

// Next advances the iterator to the next item and returns true, or returns
// false if the end of the iteration has been reached or if an error has
// occurred.
func (iter *LCPIterator) Next() bool { return iter.impl.Next() }

// Skip is equivalent to calling Next() n times, but faster.
func (iter *LCPIterator) Skip(n uint64) bool { return iter.impl.Skip(n) }

// Index returns the index of the current suffix array entry.
func (iter *LCPIterator) Index() uint64 { return iter.impl.Index() }

// Height returns the LCP of the current suffix array entry with the previous.
func (iter *LCPIterator) Height() uint64 { return iter.impl.Value() }

// SetHeight assigns the LCP for the current entry and its predecessor.
func (iter *LCPIterator) SetHeight(height uint64) { iter.impl.SetValue(height) }

// Err returns the error which caused Next() to return false.
func (iter *LCPIterator) Err() error { return iter.impl.Err() }

// Flush ensures that all pending writes have reached the OS.
func (iter *LCPIterator) Flush() error { return iter.impl.Flush() }

// Close flushes writes and frees the resources used by the iterator.
func (iter *LCPIterator) Close() error { return iter.impl.Close() }

// BuildLCPArray constructs the LCP array for the given Text and SuffixArray.
//
// (LCP = "Longest Common Prefix")
//
// Suppose we have an input text "banana".  The suffix array for that text is:
//
//   [6 5 3 1 0 4 2]
//
// Which is shorthand for:
//
//   6 $
//   5 a$
//   3 ana$
//   1 anana$
//   0 banana$
//   4 na$
//   2 nana$
//
// Each index in the LCP array represents the number of leading characters that
// are shared between the previous SA entry and the current SA entry.
//
//   ! $         <-- undefined
//   0 a$        <-- zero leading chars shared with $
//   1 ana$      <-- one leading char shared with a$
//   3 anana$    <-- three leading chars shared with ana$
//   0 banana$   <-- no common leading chars with anana$
//   0 na$       <-- no common leading chars with banana$
//   2 nana$     <-- two leading chars with na$
//
// Reference:
//
//  [1] “Linear-Time Longest-Common-Prefix Computation in Suffix Arrays and Its Applications”,
//      Toru Kasai, Gunho Lee, Hiroki Arimura, Setsuo Arikawa, and Kunsoo Park.
//      https://doi.org/10.1007/3-540-48194-X_17
//
func BuildLCPArray(text *Text, sa *SuffixArray, opts ...Option) (*LCPArray, error) {
	lcpOpts := extendOptions(
		opts,
		NumValues(sa.Len()))

	rankOpts := extendOptions(
		opts,
		NumValues(sa.Len()),
		BytesPerValue(8),
		WithFile(nil))

	// Rank array RA is the inverse of the suffix array SA.
	//
	// That is, where SA maps:
	//
	//   index of suffix in the sorted list of suffices
	//   -> text offset of that suffix
	//
	// RA instead maps each offset in the text to the corresponding index
	// in SA for the suffix starting at that offset.
	rankArray, err := makeBigArray(rankOpts)
	if err != nil {
		return nil, err
	}
	defer rankArray.Close()

	err = sa.ForEach(func(index uint64, pos uint64) error {
		return rankArray.SetValueAt(pos, index)
	})
	if err != nil {
		return nil, err
	}

	lcp, err := NewLCPArray(lcpOpts...)
	if err != nil {
		return nil, err
	}

	needClose := true
	defer func() {
		if needClose {
			lcp.Close()
		}
	}()

	// Algorithm below given by [1]

	h := uint64(0)
	err = bigarray.ForEach(rankArray, func(indexI, rank uint64) error {
		if rank <= 1 {
			return nil
		}

		indexJ, err := sa.PositionAt(rank - 1)
		if err != nil {
			return err
		}

		lengthI := text.Len() - indexI
		lengthJ := text.Len() - indexJ
		lengthMin := lengthI
		if lengthMin > lengthJ {
			lengthMin = lengthJ
		}

		iterI := text.Iterate(indexI+h, indexI+lengthMin)
		iterJ := text.Iterate(indexJ+h, indexJ+lengthMin)
		for iterI.Next() && iterJ.Next() && iterI.Symbol() == iterJ.Symbol() {
			h++
		}
		if err := iterI.Close(); err != nil {
			return err
		}
		if err := iterJ.Close(); err != nil {
			return err
		}

		err = lcp.SetHeightAt(rank, h)
		if err != nil {
			return err
		}

		if h > 0 {
			h--
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	needClose = false
	return lcp, nil
}

// BuildLCPLRArray constructs the LCP-LR array for an LCP array.
//
// An LCP-LR array is used to speed up string comparisons during binary search
// of the suffix array.
//
// Here are the suffix and LCP arrays for "banana.banana":
//
//      i  suffix         SA  LCP
//    [ 0] $              13   .
//    [ 1] .banana$        6   0
//    [ 2] a$             12   0
//    [ 3] a.banana$       5   1
//    [ 4] ana$           10   1
//    [ 5] ana.banana$     3   3
//    [ 6] anana$          8   3
//    [ 7] anana.banana$   1   5
//    [ 8] banana$         7   0
//    [ 9] banana.banana$  0   6
//    [10] na$            11   0
//    [11] na.banana$      4   2
//    [12] nana$           9   2
//    [13] nana.banana$    2   4
//
// Each value in the LCP-LR array is the lcp(A_i, A_j) for a single binary
// search comparison: the 0th index contains the first comparison, i=0 j=13;
// the 1st and 2nd indices contain the two permissible second comparisons,
// i=0 j=5 (left) and i=7 j=13 (right); the 3rd, 4th, 5th, and 6th indices
// contain the permissible third comparisons, i=0 j=1 (LL), i=3 j=5 (LR),
// i=7 j=9 (RL), and i=11 j=13 (RR), and so on for all O(log N) string
// comparisons.
//
// Note that once we have the LCP array, with its pairwise calculations of
// LCP[n] = lcp(n - 1, n), we can compute lcp(lo, hi) as
//
// lcp(lo, hi) = min(lcp(lo, mid - 1), lcp(mid - 1, mid), lcp(mid, mid + 1), lcp(mid + 1, hi))
//             = min(LCP-LR[2*n+1],    LCP[mid],          LCP[mid+1],        LCP-LR[2*n+2])
//
//      i  LCP-LR    computation
//     [0] 0         lcp( 0, 13) -> min(LCP-LR[1], LCP[ 6], LCP[ 7], LCP-LR[2]) -> min(0, 3, 5, 0)
//     [1] 0         lcp( 0,  5) -> min(LCP-LR[3], LCP[ 2], LCP[ 3], LCP-LR[4]) -> min(0, 0, 1, 1)
//     [2] 0         lcp( 7, 13) -> min(LCP-LR[5], LCP[10], LCP[11], LCP-LR[6]) -> min(0, 0, 2, 2)
//     [3] 0         lcp( 0,  1) -> LCP[1]                                      -> 0
//     [4] 1         lcp( 3,  5) -> min(LCP[ 4], LCP[ 5])                       -> min(1, 3)
//     [5] 0         lcp( 7,  9) -> min(LCP[ 8], LCP[ 9])                       -> min(0, 6)
//     [6] 2         lcp(11, 13) -> min(LCP[12], LCP[13])                       -> min(2, 4)
//
// ----------------------------------------------------------------------------
//
// What's the advantage of this?
//
// Well, it's a binary search, so at each step we have three indices:
//
//    lo  -> the current lower bound
//    mid -> the midpoint between lo and hi
//    hi  -> the current upper bound
//
// We also (conceptually) have four strings:
//
//    L -> the suffix starting at offset SA[lo]
//    M -> the suffix starting at offset SA[mid]
//    H -> the suffix starting at offset SA[hi]
//    P -> the phrase being searched for
//
// We need to compare M to P, so that we know whether the next step should be
// (lo, mid-1) or (mid+1, hi).  However, the suffix array is sorted, so as the
// search drills in toward the target suffix(es) we will increasingly encounter
// long common prefixes between M and P.
//
// If we know lcp(L, R), however, then we also know that all suffices *between*
// L and R, *including* M, have that many identical bytes at the start of each
// suffix.  And we discovered L and R in the first place because L ≤ P ≤ R.
//
// Therefore, P also shares at least that many bytes with M.  We have no need
// to check the first lcp(L, R) bytes of P vs M, because we can be certain that
// those two substrings are identical.
//
// Reference:
//
//  [2] Answer to Stack Overflow question “How do we Construct LCP-LR array from LCP array?”,
//      Abhijeet Ashok Muneshwar.
//      https://stackoverflow.com/a/38588642/4848155
//
func BuildLCPLRArray(lcp *LCPArray, opts ...Option) (bigarray.BigArray, error) {
	// Round lcp.Len() up to a power of 2
	n := uint64(math.Exp2(math.Ceil(math.Log2(float64(lcp.Len())))))

	opts = extendOptions(
		opts,
		NumValues(2*n+1),
		BytesPerValue(8))

	lcplr, err := makeBigArray(opts)
	if err != nil {
		return nil, err
	}

	needClose := true
	defer func() {
		if needClose {
			lcplr.Close()
		}
	}()

	iter := lcplr.Iterate(0, lcplr.Len())
	defer iter.Close()
	for iter.Next() {
		iter.SetValue(placeholder)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}

	_, err = buildLCPLR(lcplr, lcp, 0, 0, lcp.Len()-1)
	if err != nil {
		lcplr.Close()
		return nil, err
	}

	iter = lcplr.ReverseIterate(0, lcplr.Len())
	actualLength := lcplr.Len()
	for iter.Next() {
		if iter.Value() != placeholder {
			break
		}
		actualLength = iter.Index()
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}

	if actualLength < lcplr.Len() {
		if err := lcplr.Truncate(actualLength); err != nil {
			return nil, err
		}
	}

	needClose = false
	return lcplr, nil
}

func buildLCPLR(lcplr bigarray.BigArray, lcp *LCPArray, index, lo, hi uint64) (uint64, error) {
	if lo >= hi {
		panic(fmt.Errorf("BUG: %d >= %d", lo, hi))
	}

	var err error
	var h0 uint64
	h1 := ^uint64(0)
	h2 := ^uint64(0)
	h3 := ^uint64(0)

	delta := (hi - lo)
	if delta > 3 {
		mid := lo + delta/2
		x := 2*index + 1
		y := 2*index + 2

		h0, err = buildLCPLR(lcplr, lcp, x, lo, mid-1)
		if err != nil {
			return 0, err
		}

		h1, err = lcp.HeightAt(mid)
		if err != nil {
			return 0, err
		}

		h2, err = lcp.HeightAt(mid + 1)
		if err != nil {
			return 0, err
		}

		h3, err = buildLCPLR(lcplr, lcp, y, mid+1, hi)
		if err != nil {
			return 0, err
		}
	} else if delta == 3 {
		h0, err = lcp.HeightAt(lo + 1)
		if err != nil {
			return 0, err
		}

		h1, err = lcp.HeightAt(lo + 2)
		if err != nil {
			return 0, err
		}

		h2, err = lcp.HeightAt(lo + 3)
		if err != nil {
			return 0, err
		}
	} else if delta == 2 {
		h0, err = lcp.HeightAt(lo + 1)
		if err != nil {
			return 0, err
		}

		h1, err = lcp.HeightAt(lo + 2)
		if err != nil {
			return 0, err
		}
	} else {
		h0, err = lcp.HeightAt(lo + 1)
		if err != nil {
			return 0, err
		}
	}

	h := h0
	if h > h1 {
		h = h1
	}
	if h > h2 {
		h = h2
	}
	if h > h3 {
		h = h3
	}

	err = lcplr.SetValueAt(index, h)
	if err != nil {
		return 0, err
	}

	return h, nil
}
