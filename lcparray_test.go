package suffixarray

import (
	"bytes"
	"fmt"
	"testing"
)

func NaiveBuildLCPArray(text *Text, sa *SuffixArray) ([]uint64, string) {
	var buf bytes.Buffer
	lcp := make([]uint64, sa.Len())
	lcp[0] = ^uint64(0)
	buf.WriteString("[.")
	for i := 1; i < len(lcp); i++ {
		pos0, err := sa.PositionAt(uint64(i - 1))
		if err != nil {
			panic(err)
		}

		pos1, err := sa.PositionAt(uint64(i))
		if err != nil {
			panic(err)
		}

		iter0 := text.Iterate(pos0, text.Len())
		defer iter0.Close()

		iter1 := text.Iterate(pos1, text.Len())
		defer iter1.Close()

		h := uint64(0)
		for iter0.Next() && iter1.Next() && iter0.Symbol() == iter1.Symbol() {
			h++
		}
		lcp[i] = h
		fmt.Fprintf(&buf, " %d", h)
	}
	buf.WriteString("]")
	return lcp, buf.String()
}

type naiveLCPLRBuilder struct {
	text *Text
	sa   *SuffixArray
	out  []uint64
}

func (builder *naiveLCPLRBuilder) computeLCP(i, j uint64) uint64 {
	pos0, err := builder.sa.PositionAt(i)
	if err != nil {
		panic(err)
	}

	pos1, err := builder.sa.PositionAt(j)
	if err != nil {
		panic(err)
	}

	iter0 := builder.text.Iterate(pos0, builder.text.Len())
	defer iter0.Close()

	iter1 := builder.text.Iterate(pos1, builder.text.Len())
	defer iter1.Close()

	h := uint64(0)
	for iter0.Next() && iter1.Next() && iter0.Symbol() == iter1.Symbol() {
		h++
	}
	return h
}

func (builder *naiveLCPLRBuilder) recurse(index, lo, hi uint64) uint64 {
	if lo >= hi {
		panic(fmt.Errorf("BUG: %d >= %d", lo, hi))
	}

	var h0 uint64
	h1 := ^uint64(0)
	h2 := ^uint64(0)
	h3 := ^uint64(0)

	delta := (hi - lo)

	if delta > 3 {
		mid := lo + delta/2
		x := 2*index + 1
		y := 2*index + 2

		h0 = builder.recurse(x, lo, mid-1)
		h1 = builder.computeLCP(mid-1, mid)
		h2 = builder.computeLCP(mid, mid+1)
		h3 = builder.recurse(y, mid+1, hi)
	} else if delta == 3 {
		h0 = builder.computeLCP(lo, lo+1)
		h1 = builder.computeLCP(lo+1, lo+2)
		h2 = builder.computeLCP(lo+2, lo+3)
	} else if delta == 2 {
		h0 = builder.computeLCP(lo, lo+1)
		h1 = builder.computeLCP(lo+1, lo+2)
	} else {
		h0 = builder.computeLCP(lo, lo+1)
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

	hh := builder.computeLCP(lo, hi)
	if h != hh {
		panic(fmt.Errorf("BUG: h0=%d h1=%d h2=%d h3=%d h=%d hh=%d", h0, h1, h2, h3, h, hh))
	}

	builder.out[index] = hh
	return hh
}

func NaiveBuildLCPLRArray(text *Text, sa *SuffixArray) ([]uint64, string) {
	n := uint64(1)
	for sa.Len() > n {
		n *= 2
	}

	lcplr := make([]uint64, 2*n+1)
	for i := range lcplr {
		lcplr[i] = ^uint64(0)
	}

	builder := &naiveLCPLRBuilder{
		text: text,
		sa:   sa,
		out:  lcplr,
	}
	builder.recurse(0, 0, sa.Len()-1)

	finalLen := uint64(len(lcplr))
	for finalLen > 0 {
		if lcplr[finalLen-1] != ^uint64(0) {
			break
		}
		finalLen--
	}
	lcplr = lcplr[0:finalLen]

	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, h := range lcplr {
		if i > 0 {
			buf.WriteByte(' ')
		}
		if h == ^uint64(0) {
			buf.WriteByte('.')
		} else {
			fmt.Fprintf(&buf, "%d", h)
		}
	}
	buf.WriteByte(']')

	return lcplr, buf.String()
}

func TestBuildLCPArray(t *testing.T) {
	type testrow struct {
		Input    string
		SAInput  string
		Expected string
	}
	for _, cfg := range configurations {
		t.Logf("starting %s tests", cfg.Name)
		opts := cfg.Opts

		for i, row := range []testrow{
			testrow{banana, bananaSA, bananaLCP},
			testrow{banana2, banana2SA, banana2LCP},
			testrow{cabbage, cabbageSA, cabbageLCP},
			testrow{loremIpsum, loremIpsumSA, loremIpsumLCP},
			testrow{abcdefgh, abcdefghSA, abcdefghLCP},
			testrow{aaaaaaaa, aaaaaaaaSA, aaaaaaaaLCP},
		} {
			text := NewTextFromString(row.Input, opts...)
			sa := NewFromString(row.SAInput, opts...)

			lcp, err := BuildLCPArray(text, sa, opts...)
			if err != nil {
				t.Errorf("[%s/%03d] BuildLCPArray %q, %v: error: %v", cfg.Name, i, row.Input, sa.Debug(), err)
				continue
			}

			actual := lcp.Debug()
			if row.Expected != actual {
				t.Errorf("[%s/%03d] BuildLCPArray %q, %v: expected %v, got %v", cfg.Name, i, row.Input, sa.Debug(), row.Expected, actual)
			}

			_, naivestr := NaiveBuildLCPArray(text, sa)
			if row.Expected != naivestr {
				t.Errorf("[%s/%03d] BuildLCPArray %q, %v: expected %v, but should expect %v", cfg.Name, i, row.Input, sa.Debug(), row.Expected, naivestr)
			}
		}
	}
}

func TestBuildLCPLRArray(t *testing.T) {
	type testrow struct {
		Input    string
		SAInput  string
		LCPInput string
		Expected string
	}
	for _, cfg := range configurations {
		t.Logf("starting %s tests", cfg.Name)
		opts := cfg.Opts

		for i, row := range []testrow{
			testrow{banana, bananaSA, bananaLCP, bananaLCPLR},
			testrow{banana2, banana2SA, banana2LCP, banana2LCPLR},
			testrow{cabbage, cabbageSA, cabbageLCP, cabbageLCPLR},
			testrow{loremIpsum, loremIpsumSA, loremIpsumLCP, loremIpsumLCPLR},
			testrow{abcdefgh, abcdefghSA, abcdefghLCP, abcdefghLCPLR},
			testrow{aaaaaaaa, aaaaaaaaSA, aaaaaaaaLCP, aaaaaaaaLCPLR},
		} {
			text := NewTextFromString(row.Input, opts...)
			sa := NewFromString(row.SAInput, opts...)
			lcp := NewLCPArrayFromString(row.LCPInput, opts...)

			lcplr, err := BuildLCPLRArray(lcp, opts...)
			if err != nil {
				t.Errorf("[%s/%03d] BuildLCPLRArray %s: error: %v", cfg.Name, i, lcp.Debug(), err)
			}

			actual := lcplr.Debug()
			if row.Expected != actual {
				t.Errorf("[%s/%03d] BuildLCPLRArray %q, %v: expected %v, got %v", cfg.Name, i, row.Input, sa.Debug(), row.Expected, actual)
			}

			_, naivestr := NaiveBuildLCPLRArray(text, sa)
			if row.Expected != naivestr {
				t.Errorf("[%s/%03d] BuildLCPLRArray %q, %v: expected %v, but should expect %v", cfg.Name, i, row.Input, sa.Debug(), row.Expected, naivestr)
			}
		}
	}
}
