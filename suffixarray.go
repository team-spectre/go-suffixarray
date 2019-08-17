package suffixarray

import (
	bigarray "github.com/team-spectre/go-bigarray"
)

// SuffixArray represents a suffix array.
//
// A suffix array is a sequence of offsets into a text, with each offset N
// standing in for the string TEXT[N:], i.e. a suffix of TEXT.  The suffixes
// are sorted lexicographically, with SA[0] always being the empty suffix at
// the end of the text.
//
type SuffixArray struct {
	ba bigarray.BigArray
}

// Iterator iterates through a SuffixArray.
type Iterator struct {
	impl bigarray.Iterator
}

// New constructs a new SuffixArray.
func New(opts ...Option) (*SuffixArray, error) {
	ba, err := makeBigArray(opts)
	if err != nil {
		return nil, err
	}
	sa := &SuffixArray{ba}
	return sa, nil
}

// MaxValue returns the maximum offset which can be stored in the array.
func (sa *SuffixArray) MaxValue() uint64 { return sa.ba.MaxValue() }

// Len returns the length of the array, which is always one greater than the
// length of the text.
func (sa *SuffixArray) Len() uint64 { return sa.ba.Len() }

// PositionAt returns the text offset of the suffix at the given index.
// The idea is that TEXT[SA[index-1]:] < TEXT[SA[index]:] < TEXT[SA[index+1]:].
//
func (sa *SuffixArray) PositionAt(index uint64) (uint64, error) {
	return sa.ba.ValueAt(index)
}

// SetPositionAt sets the text offset of the suffix at the given index.
func (sa *SuffixArray) SetPositionAt(index uint64, pos uint64) error {
	return sa.ba.SetValueAt(index, pos)
}

// Iterate constructs an Iterator in the forward direction.
func (sa *SuffixArray) Iterate(i, j uint64) *Iterator {
	return &Iterator{impl: sa.ba.Iterate(i, j)}
}

// ReverseIterate constructs an Iterator in the reverse direction.
func (sa *SuffixArray) ReverseIterate(i, j uint64) *Iterator {
	return &Iterator{impl: sa.ba.ReverseIterate(i, j)}
}

// ForEach is a convenience method that forward-iterates over the entire array.
func (sa *SuffixArray) ForEach(fn func(uint64, uint64) error) error {
	return bigarray.ForEach(sa.ba, fn)
}

// ReverseForEach is a convenience method that reverse-iterates over the entire array.
func (sa *SuffixArray) ReverseForEach(fn func(uint64, uint64) error) error {
	return bigarray.ReverseForEach(sa.ba, fn)
}

// CopyFrom copies the text offsets from another SuffixArray to this one.  The
// two arrays must have the same Len(), and this text must have a MaxValue()
// large enough to accommodate any offset in the source array.
func (sa *SuffixArray) CopyFrom(src *SuffixArray) error {
	return sa.ba.CopyFrom(src.ba)
}

// Truncate trims the array to the given length.
func (sa *SuffixArray) Truncate(length uint64) error { return sa.ba.Truncate(length) }

// Freeze makes the array read-only.
func (sa *SuffixArray) Freeze() error { return sa.ba.Freeze() }

// Flush ensures that all pending writes have reached the OS.
func (sa *SuffixArray) Flush() error { return sa.ba.Flush() }

// Close flushes any writes and frees the resources used by the array.
func (sa *SuffixArray) Close() error { return sa.ba.Close() }

// Debug returns a human-friendly debugging representation of the array.
func (sa *SuffixArray) Debug() string { return sa.ba.Debug() }

// Clear overwrites the array with the placeholder value 2^64-1.
func (sa *SuffixArray) Clear() error {
	iter := sa.Iterate(0, sa.Len())
	for iter.Next() {
		iter.SetPosition(placeholder)
	}
	return iter.Close()
}

// Next advances the iterator to the next item and returns true, or returns
// false if the end of the iteration has been reached or if an error has
// occurred.
func (iter *Iterator) Next() bool { return iter.impl.Next() }

// Skip is equivalent to calling Next() n times, but faster.
func (iter *Iterator) Skip(n uint64) bool { return iter.impl.Skip(n) }

// Index returns the index of the current suffix.
func (iter *Iterator) Index() uint64 { return iter.impl.Index() }

// Position returns the text offset of the current suffix.
func (iter *Iterator) Position() uint64 { return iter.impl.Value() }

// SetPosition replaces the text offset of the current suffix.
func (iter *Iterator) SetPosition(pos uint64) { iter.impl.SetValue(pos) }

// Err returns the error which caused Next() to return false.
func (iter *Iterator) Err() error { return iter.impl.Err() }

// Flush ensures that all pending writes have reached the OS.
func (iter *Iterator) Flush() error { return iter.impl.Flush() }

// Close flushes writes and frees the resources used by the iterator.
func (iter *Iterator) Close() error { return iter.impl.Close() }
