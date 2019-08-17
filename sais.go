package suffixarray

import (
	bigarray "github.com/team-spectre/go-bigarray"
)

const placeholder = ^uint64(0)

func guessLMSSort(text *Text, typeMap *TypeMap, bucketSizes []uint64, opts []Option) (*SuffixArray, error) {
	opts = extendOptions(
		opts,
		NumValues(text.Len()+1),
		MaxValue(placeholder))

	sa, err := New(opts...)
	if err != nil {
		return nil, err
	}

	err = sa.Clear()
	if err != nil {
		return nil, err
	}

	needClose := true
	defer func() {
		if needClose {
			sa.Close()
		}
	}()

	if err := sa.SetPositionAt(0, text.Len()); err != nil {
		return nil, err
	}

	textIter := text.Iterate(0, text.Len())
	typeIter := typeMap.Iterate(0, text.Len())
	bucketIters := BuildBucketTailIterators(bucketSizes, sa)

	needIterClose := true
	defer func() {
		if needIterClose {
			textIter.Close()
			typeIter.Close()
			for _, iter := range bucketIters {
				iter.Close()
			}
		}
	}()

	for textIter.Next() && typeIter.Next() {
		if typeIter.IsLMS() {
			tail := bucketIters[textIter.Symbol()]
			if !tail.Next() {
				break
			}
			tail.SetPosition(textIter.Index())
		}
	}

	if err := textIter.Close(); err != nil {
		return nil, err
	}
	if err := typeIter.Close(); err != nil {
		return nil, err
	}
	for _, iter := range bucketIters {
		if err := iter.Close(); err != nil {
			return nil, err
		}
	}

	needIterClose = false
	needClose = false
	return sa, nil
}

func induceSortL(text *Text, typeMap *TypeMap, bucketSizes []uint64, sa *SuffixArray) error {
	saIter := sa.Iterate(0, sa.Len())
	bucketIters := BuildBucketHeadIterators(bucketSizes, sa)

	needClose := true
	defer func() {
		if needClose {
			saIter.Close()
			for _, iter := range bucketIters {
				iter.Close()
			}
		}
	}()

	for saIter.Next() {
		if saIter.Position() == 0 || saIter.Position() == placeholder {
			continue
		}
		j := saIter.Position() - 1

		bit, err := typeMap.TypeAt(j)
		if err != nil {
			return err
		}
		if bit == SType {
			continue
		}

		symbol, err := text.SymbolAt(j)
		if err != nil {
			return err
		}

		iter := bucketIters[symbol]
		if !iter.Next() {
			break
		}
		iter.SetPosition(j)
	}
	if err := saIter.Close(); err != nil {
		return err
	}
	for _, iter := range bucketIters {
		if err := iter.Close(); err != nil {
			return err
		}
	}

	needClose = false
	return nil
}

func induceSortR(text *Text, typeMap *TypeMap, bucketSizes []uint64, sa *SuffixArray) error {
	saIter := sa.ReverseIterate(0, sa.Len())
	bucketIters := BuildBucketTailIterators(bucketSizes, sa)

	needClose := true
	defer func() {
		if needClose {
			saIter.Close()
			for _, iter := range bucketIters {
				iter.Close()
			}
		}
	}()

	for saIter.Next() {
		if saIter.Position() == 0 || saIter.Position() == placeholder {
			continue
		}
		j := saIter.Position() - 1

		bit, err := typeMap.TypeAt(j)
		if err != nil {
			return err
		}
		if bit == LType {
			continue
		}

		symbol, err := text.SymbolAt(j)
		if err != nil {
			return err
		}

		iter := bucketIters[symbol]
		if !iter.Next() {
			break
		}
		iter.SetPosition(j)
	}
	if err := saIter.Close(); err != nil {
		return err
	}
	for _, iter := range bucketIters {
		if err := iter.Close(); err != nil {
			return err
		}
	}

	needClose = false
	return nil
}

func summarize(text *Text, typeMap *TypeMap, sa *SuffixArray, opts []Option) (*Text, bigarray.BigArray, error) {
	lmsOpts := extendOptions(
		opts,
		NumValues(text.Len()+1),
		MaxValue(placeholder))

	lmsNames, err := New(lmsOpts...)
	if err != nil {
		return nil, nil, err
	}
	defer lmsNames.Close()

	err = lmsNames.Clear()
	if err != nil {
		return nil, nil, err
	}

	var currentName uint64
	var lastLMSSuffixOffset uint64
	err = sa.ForEach(func(index uint64, pos uint64) error {
		if index == 0 {
			lastLMSSuffixOffset = pos
			return lmsNames.SetPositionAt(pos, currentName)
		}
		lms, err := typeMap.IsLMS(pos)
		if err != nil {
			return err
		}
		if !lms {
			return nil
		}
		eq, err := typeMap.LMSSubstringsAreEqual(text, lastLMSSuffixOffset, pos)
		if err != nil {
			return err
		}
		if !eq {
			currentName++
		}
		lastLMSSuffixOffset = pos
		return lmsNames.SetPositionAt(pos, currentName)
	})
	if err != nil {
		return nil, nil, err
	}

	maxValue := currentName
	if maxValue == 0 {
		maxValue++
	}

	summaryTextOpts := extendOptions(
		opts,
		NumValues(lmsNames.Len()),
		MaxValue(maxValue))

	summaryText, err := NewText(currentName+1, summaryTextOpts...)
	if err != nil {
		return nil, nil, err
	}

	needSummaryTextClose := true
	defer func() {
		if needSummaryTextClose {
			summaryText.Close()
		}
	}()

	opts = extendOptions(
		opts,
		BytesPerValue(8),
		NumValues(lmsNames.Len()))

	summarySuffixOffsets, err := makeBigArray(opts)
	if err != nil {
		return nil, nil, err
	}

	needSummarySuffixOffsetsClose := true
	defer func() {
		if needSummarySuffixOffsetsClose {
			summarySuffixOffsets.Close()
		}
	}()

	summaryTextActualLen := uint64(0)
	summarySuffixOffsetsActualLen := uint64(0)
	err = lmsNames.ForEach(func(index uint64, name uint64) error {
		if name == placeholder {
			return nil
		}
		if err := summaryText.SetSymbolAt(summaryTextActualLen, name); err != nil {
			return err
		}
		summaryTextActualLen++
		if err := summarySuffixOffsets.SetValueAt(summarySuffixOffsetsActualLen, index); err != nil {
			return err
		}
		summarySuffixOffsetsActualLen++
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	err = summaryText.Truncate(summaryTextActualLen)
	if err != nil {
		return nil, nil, err
	}

	err = summarySuffixOffsets.Truncate(summarySuffixOffsetsActualLen)
	if err != nil {
		return nil, nil, err
	}

	needSummaryTextClose = false
	needSummarySuffixOffsetsClose = false
	return summaryText, summarySuffixOffsets, nil
}

func buildSummarySuffixArray(summaryText *Text, opts []Option) (*SuffixArray, error) {
	if summaryText.Len() == summaryText.AlphabetSize() {
		opts = extendOptions(
			opts,
			NumValues(summaryText.Len()+1),
			MaxValue(summaryText.Len()))

		ssa, err := New(opts...)
		if err != nil {
			return nil, err
		}

		needClose := true
		defer func() {
			if needClose {
				ssa.Close()
			}
		}()

		err = ssa.SetPositionAt(0, summaryText.Len())
		if err != nil {
			return nil, err
		}

		err = summaryText.ForEach(func(stringIndex uint64, symbol uint64) error {
			return ssa.SetPositionAt(symbol+1, stringIndex)
		})
		if err != nil {
			return nil, err
		}

		needClose = false
		return ssa, nil
	}
	return BuildSuffixArray(summaryText, opts...)
}

func exactLMSSort(text *Text, typeMap *TypeMap, bucketSizes []uint64, summarySuffixArray *SuffixArray, summarySuffixOffsets bigarray.BigArray, opts []Option) (*SuffixArray, error) {
	opts = extendOptions(
		opts,
		NumValues(text.Len()+1),
		MaxValue(placeholder))

	sa, err := New(opts...)
	if err != nil {
		return nil, err
	}

	needClose := true
	defer func() {
		if needClose {
			sa.Close()
		}
	}()

	err = sa.Clear()
	if err != nil {
		return nil, err
	}

	err = sa.SetPositionAt(0, text.Len())
	if err != nil {
		return nil, err
	}

	bucketTails := BuildBucketTails(bucketSizes)
	err = summarySuffixArray.ReverseForEach(func(index uint64, metapos uint64) error {
		if index == 0 || index == 1 {
			return nil
		}
		stringIndex, err := summarySuffixOffsets.ValueAt(metapos)
		if err != nil {
			return err
		}
		sym, err := text.SymbolAt(stringIndex)
		if err != nil {
			return err
		}
		err = sa.SetPositionAt(bucketTails[sym], stringIndex)
		bucketTails[sym]--
		return err
	})
	if err != nil {
		return nil, err
	}

	needClose = false
	return sa, nil
}

// BuildSuffixArray constructs the suffix array for the given text, using the
// SA-IS algorithm.
//
// References:
//  [1] “A walk through the SA-IS Suffix Array Construction Algorithm”,
//      http://zork.net/~st/jottings/sais.html
//
func BuildSuffixArray(text *Text, opts ...Option) (*SuffixArray, error) {
	if text.Len() == 0 {
		opts = extendOptions(
			opts,
			NumValues(1),
			BytesPerValue(1),
			MaxValue(0))

		sa, err := New(opts...)
		if err != nil {
			return nil, err
		}

		err = sa.SetPositionAt(0, 0)
		if err != nil {
			sa.Close()
			return nil, err
		}
		return sa, nil
	}

	if text.Len() == 1 {
		opts = extendOptions(
			opts,
			NumValues(2),
			BytesPerValue(1),
			MaxValue(0))

		sa, err := New(opts...)
		if err != nil {
			return nil, err
		}

		values := []uint64{1, 0}
		iter := sa.Iterate(0, 2)
		for iter.Next() {
			iter.SetPosition(values[iter.Index()])
		}
		err = iter.Close()
		if err != nil {
			sa.Close()
			return nil, err
		}
		return sa, nil
	}

	typeMap, err := BuildTypeMap(text, opts...)
	if err != nil {
		return nil, err
	}
	defer typeMap.Close()

	bucketSizes, err := BuildBucketSizes(text)
	if err != nil {
		return nil, err
	}

	guessed, err := guessLMSSort(text, typeMap, bucketSizes, opts)
	if err != nil {
		return nil, err
	}
	defer guessed.Close()

	err = induceSortL(text, typeMap, bucketSizes, guessed)
	if err != nil {
		return nil, err
	}

	err = induceSortR(text, typeMap, bucketSizes, guessed)
	if err != nil {
		return nil, err
	}

	summaryText, summarySuffixOffsets, err := summarize(text, typeMap, guessed, opts)
	if err != nil {
		return nil, err
	}
	defer summaryText.Close()
	defer summarySuffixOffsets.Close()

	summarySuffixArray, err := buildSummarySuffixArray(summaryText, opts)
	if err != nil {
		return nil, err
	}
	defer summarySuffixArray.Close()

	exact, err := exactLMSSort(text, typeMap, bucketSizes, summarySuffixArray, summarySuffixOffsets, opts)
	if err != nil {
		return nil, err
	}

	err = induceSortL(text, typeMap, bucketSizes, exact)
	if err != nil {
		return nil, err
	}

	err = induceSortR(text, typeMap, bucketSizes, exact)
	if err != nil {
		return nil, err
	}

	return exact, nil
}
