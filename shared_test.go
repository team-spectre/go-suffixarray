package suffixarray

import (
	"strconv"
	"strings"
	"sync"

	bigarray "github.com/team-spectre/go-bigarray"
)

const banana = `banana`
const bananaTM = `[L@L@LL@]`
const bananaSA = `[6 5 3 1 0 4 2]`
const bananaLCP = `[. 0 1 3 0 0 2]`
const bananaLCPLR = `[0 0 0]`

const banana2 = `banana.banana`
const banana2TM = `[L@L@LL@L@L@LL@]`
const banana2SA = `[13 6 12 5 10 3 8 1 7 0 11 4 9 2]`
const banana2LCP = `[. 0 0 1 1 3 3 5 0 6 0 2 2 4]`
const banana2LCPLR = `[0 0 0 0 1 0 2]`

const cabbage = `cabbage`
const cabbageTM = `[L@LL@LL@]`
const cabbageSA = `[7 1 4 3 2 0 6 5]`
const cabbageLCP = `[. 0 1 0 1 0 0 0]`
const cabbageLCPLR = `[0 0 0]`

const loremIpsum = `Lorem ipsum dolor sit amet, consectetur adipiscing elit.`
const loremIpsumTM = `[SSL@L@SSSLL@SL@SL@L@L@SL@LL@SL@LL@L@SLL@SSSL@L@SLL@SL@LL@]`
const loremIpsumSA = `[56 39 21 27 11 50 5 17 26 55 0 40 22 46 28 33 41 12 32 51 3 24 35 49 47 42 6 44 19 53 52 14 10 4 23 48 30 13 29 15 1 43 7 38 16 2 45 31 18 8 20 25 54 34 36 9 37]`
const loremIpsumLCP = `[. 0 2 1 1 1 1 1 0 0 0 0 1 0 1 1 0 1 0 1 1 1 2 0 0 1 2 1 1 2 0 1 0 2 1 0 1 0 1 1 2 0 1 0 2 1 0 1 1 1 0 1 1 1 1 0 1]`
const loremIpsumLCPLR = `[0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 1 0 0 1 0 1 0 0 2 1 1 0]`

const abcdefgh = `abcdefgh`
const abcdefghTM = `[SSSSSSSL@]`
const abcdefghSA = `[8 0 1 2 3 4 5 6 7]`
const abcdefghLCP = `[. 0 0 0 0 0 0 0 0]`
const abcdefghLCPLR = `[0 0 0]`

const aaaaaaaa = `aaaaaaaa`
const aaaaaaaaTM = `[LLLLLLLL@]`
const aaaaaaaaSA = `[8 7 6 5 4 3 2 1 0]`
const aaaaaaaaLCP = `[. 0 1 2 3 4 5 6 7]`
const aaaaaaaaLCPLR = `[0 0 5]`

func NewTextFromString(str string, opts ...Option) *Text {
	opts = extendOptions(
		opts,
		NumValues(uint64(len(str))),
		BytesPerValue(1),
		WithReadOnlyFile(strings.NewReader(str)))

	text, err := NewText(256, opts...)
	if err != nil {
		panic(err)
	}

	return text
}

func NewBigArrayFromString(str string, opts ...Option) bigarray.BigArray {
	pieces := strings.Split(str[1:len(str)-1], " ")

	opts = extendOptions(
		opts,
		NumValues(uint64(len(pieces))),
		BytesPerValue(8))

	ba, err := makeBigArray(opts)
	if err != nil {
		panic(err)
	}

	iter := ba.Iterate(0, ba.Len())
	for iter.Next() {
		value := ^uint64(0)
		piece := pieces[iter.Index()]
		if piece != "." {
			value, err = strconv.ParseUint(pieces[iter.Index()], 10, 64)
			if err != nil {
				panic(err)
			}
		}
		iter.SetValue(value)
	}
	err = iter.Close()
	if err != nil {
		panic(err)
	}

	return ba
}

func NewFromString(str string, opts ...Option) *SuffixArray {
	pieces := strings.Split(str[1:len(str)-1], " ")

	opts = extendOptions(
		opts,
		NumValues(uint64(len(pieces))),
		MaxValue(^uint64(0)))

	sa, err := New(opts...)
	if err != nil {
		panic(err)
	}

	iter := sa.Iterate(0, sa.Len())
	for iter.Next() {
		value := ^uint64(0)
		piece := pieces[iter.Index()]
		if piece != "." {
			value, err = strconv.ParseUint(pieces[iter.Index()], 10, 64)
			if err != nil {
				panic(err)
			}
		}
		iter.SetPosition(value)
	}
	err = iter.Close()
	if err != nil {
		panic(err)
	}

	return sa
}

func NewLCPArrayFromString(str string, opts ...Option) *LCPArray {
	pieces := strings.Split(str[1:len(str)-1], " ")

	opts = extendOptions(
		opts,
		NumValues(uint64(len(pieces))))

	lcp, err := NewLCPArray(opts...)
	if err != nil {
		panic(err)
	}

	iter := lcp.Iterate(0, lcp.Len())
	for iter.Next() {
		value := ^uint64(0)
		piece := pieces[iter.Index()]
		if piece != "." {
			value, err = strconv.ParseUint(pieces[iter.Index()], 10, 64)
			if err != nil {
				panic(err)
			}
		}
		iter.SetHeight(value)
	}
	err = iter.Close()
	if err != nil {
		panic(err)
	}

	return lcp
}

type configuration struct {
	Name string
	Opts []Option
}

var globalPool = &sync.Pool{
	New: func() interface{} {
		return make([]byte, 1048576)
	},
}

var configurations = []configuration{
	configuration{
		Name: "mem",
		Opts: nil,
	},
	configuration{
		Name: "disk",
		Opts: []Option{
			OnDiskThreshold(0),
		},
	},
	configuration{
		Name: "disk+pool",
		Opts: []Option{
			OnDiskThreshold(0),
			WithPool(globalPool),
		},
	},
}
