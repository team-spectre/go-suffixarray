package suffixarray

import (
	bigarray "github.com/team-spectre/go-bigarray"
	bigbitvector "github.com/team-spectre/go-bigbitvector"
)

func extendOptions(original []Option, more ...Option) []Option {
	m := uint(len(original))
	n := uint(len(more))
	dupe := make([]Option, m+n)
	copy(dupe[0:m], original[0:m])
	copy(dupe[m:m+n], more[0:n])
	return dupe
}

func makeBigArray(list []Option) (bigarray.BigArray, error) {
	out := make([]bigarray.Option, 0, len(list))
	for _, item := range list {
		if item.BigArrayOption != nil {
			out = append(out, item.BigArrayOption)
		}
	}
	return bigarray.New(out...)
}

func makeBigBitVector(list []Option) (bigbitvector.BigBitVector, error) {
	out := make([]bigbitvector.Option, 0, len(list))
	for _, item := range list {
		if item.BigBitVectorOption != nil {
			out = append(out, item.BigBitVectorOption)
		}
	}
	return bigbitvector.New(out...)
}
