package suffixarray

// BuildBucketSizes scans the text, counting the number of occurrences of each
// symbol, and returns an array of the counts (indexed by symbol).
func BuildBucketSizes(text *Text) ([]uint64, error) {
	out := make([]uint64, text.AlphabetSize())
	err := text.ForEach(func(_ uint64, symbol uint64) error {
		out[symbol]++
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BuildBucketHeads returns an array of starting indices in the suffix array
// for each symbol.
func BuildBucketHeads(bucketSizes []uint64) []uint64 {
	out := make([]uint64, len(bucketSizes))
	offset := uint64(1)
	for index, count := range bucketSizes {
		out[index] = offset
		offset += count
	}
	return out
}

// BuildBucketTails returns an array of starting indices in the suffix array
// for each symbol.
func BuildBucketTails(bucketSizes []uint64) []uint64 {
	out := make([]uint64, len(bucketSizes))
	offset := uint64(1)
	for index, count := range bucketSizes {
		offset += count
		out[index] = offset - 1
	}
	return out
}

// BuildBucketHeadIterators creates a forward iterator for each symbol,
// iterating over the symbol's indices in the suffix array head-first.
func BuildBucketHeadIterators(bucketSizes []uint64, sa *SuffixArray) []*Iterator {
	out := make([]*Iterator, len(bucketSizes))
	offset := uint64(1)
	for index, count := range bucketSizes {
		out[index] = sa.Iterate(offset, offset+count)
		offset += count
	}
	return out
}

// BuildBucketTailIterators creates a reverse iterator for each symbol,
// iterating over the symbol's indices in the suffix array tail-first.
func BuildBucketTailIterators(bucketSizes []uint64, sa *SuffixArray) []*Iterator {
	out := make([]*Iterator, len(bucketSizes))
	offset := uint64(1)
	for index, count := range bucketSizes {
		out[index] = sa.ReverseIterate(offset, offset+count)
		offset += count
	}
	return out
}
