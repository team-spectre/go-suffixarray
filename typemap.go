package suffixarray

import (
	"bytes"

	bigbitvector "github.com/team-spectre/go-bigbitvector"
)

const (
	// SType indicates that an offset in the corresponding text contains an
	// "S-type" symbol, in the terminology of the SA-IS algorithm.
	SType = false

	// LType indicates that an offset in the corresponding text contains an
	// "L-type" symbol, in the terminology of the SA-IS algorithm.
	LType = true
)

// TypeMap catalogues the SA-IS type of each symbol in a text.
type TypeMap struct {
	bv bigbitvector.BigBitVector
}

// TypeMapIterator iterates over a TypeMap.
//
// The basic usage pattern is:
//
//   iter := typeMap.Iterate(i, j)
//   for iter.Next() {
//     ... // call Index(), Type(), IsLMS(), and/or SetType()
//   }
//   err := iter.Close()
//   if err != nil {
//     ... // handle error
//   }
//
// Iterators are created in an indeterminate state; the caller must invoke
// Next() to advance to the first item.
//
type TypeMapIterator struct {
	impl   bigbitvector.Iterator
	last   bool
	primed bool
}

// NewTypeMap constructs a new TypeMap.
func NewTypeMap(opts ...Option) (*TypeMap, error) {
	bv, err := makeBigBitVector(opts)
	if err != nil {
		return nil, err
	}
	return &TypeMap{bv}, nil
}

// Len returns the number of symbol types in the TypeMap.
func (typeMap *TypeMap) Len() uint64 { return typeMap.bv.Len() }

// TypeAt returns the type for the symbol at index.
func (typeMap *TypeMap) TypeAt(index uint64) (bool, error) {
	return typeMap.bv.BitAt(index)
}

// SetTypeAt replaces the type for the symbol at index.
func (typeMap *TypeMap) SetTypeAt(index uint64, typeBit bool) error {
	return typeMap.bv.SetBitAt(index, typeBit)
}

// Iterate constructs a TypeMapIterator in the forward direction.
func (typeMap *TypeMap) Iterate(i, j uint64) *TypeMapIterator {
	return &TypeMapIterator{impl: typeMap.bv.Iterate(i, j)}
}

// ForEach is a convenience method that iterates over the entire TypeMap.
func (typeMap *TypeMap) ForEach(fn func(uint64, bool, bool) error) error {
	iter := typeMap.Iterate(0, typeMap.Len())
	defer iter.Close()
	for iter.Next() {
		err := fn(iter.Index(), iter.Type(), iter.IsLMS())
		if err != nil {
			return err
		}
	}
	return iter.Close()
}

// CopyFrom copies the symbol types from another TypeMap to this one.  The two
// TypeMaps must have the same Len().
func (typeMap *TypeMap) CopyFrom(src *TypeMap) error {
	return typeMap.bv.CopyFrom(src.bv)
}

// Truncate trims the TypeMap to the given length.
func (typeMap *TypeMap) Truncate(length uint64) error { return typeMap.bv.Truncate(length) }

// Freeze makes the TypeMap read-only.
func (typeMap *TypeMap) Freeze() error { return typeMap.bv.Freeze() }

// Flush ensures that all pending writes have reached the OS.
func (typeMap *TypeMap) Flush() error { return typeMap.bv.Flush() }

// Close flushes any writes and frees the resources used by the TypeMap.
func (typeMap *TypeMap) Close() error { return typeMap.bv.Close() }

// IsLMS returns true iff the symbol at the given offset is a "left-most S".
func (typeMap *TypeMap) IsLMS(index uint64) (bool, error) {
	if index < 1 {
		return false, nil
	}
	a, err := typeMap.bv.BitAt(index - 1)
	if err != nil {
		return false, err
	}
	b, err := typeMap.bv.BitAt(index)
	if err != nil {
		return false, err
	}
	return b == SType && a == LType, nil
}

// LMSSubstringsAreEqual returns true iff the two LMS-terminated strings at
// offsets i and j are equal.
func (typeMap *TypeMap) LMSSubstringsAreEqual(text *Text, i, j uint64) (bool, error) {
	n := text.Len()
	if i >= n || j >= n {
		return i == j, nil
	}
	k := uint64(0)
	for {
		lmsI, err := typeMap.IsLMS(i + k)
		if err != nil {
			return false, err
		}
		lmsJ, err := typeMap.IsLMS(j + k)
		if err != nil {
			return false, err
		}

		if k > 0 && lmsI && lmsJ {
			return true, nil
		}
		if lmsI != lmsJ {
			return false, nil
		}

		symI, err := text.SymbolAt(i + k)
		if err != nil {
			return false, err
		}
		symJ, err := text.SymbolAt(j + k)
		if err != nil {
			return false, err
		}

		if symI != symJ {
			return false, nil
		}
		k++
	}
}

// Debug returns a human-friendly debugging representation of the TypeMap.
func (typeMap *TypeMap) Debug() string {
	var buf bytes.Buffer
	var last bool
	buf.WriteByte('[')
	err := bigbitvector.ForEach(typeMap.bv, func(index uint64, bit bool) error {
		if index > 0 && last == LType && bit == SType {
			buf.WriteByte('@')
		} else if bit == SType {
			buf.WriteByte('S')
		} else {
			buf.WriteByte('L')
		}
		last = bit
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
func (iter *TypeMapIterator) Next() bool { return iter.Skip(1) }

// Index returns the index of the current symbol.
func (iter *TypeMapIterator) Index() uint64 { return iter.impl.Index() }

// Type returns the type of the current symbol.
func (iter *TypeMapIterator) Type() bool { return iter.impl.Bit() }

// IsLMS returns true iff the current symbol is a "left-most S".
func (iter *TypeMapIterator) IsLMS() bool { return iter.last == LType && iter.Type() == SType }

// SetType replaces the type for the current symbol.
func (iter *TypeMapIterator) SetType(bit bool) { iter.impl.SetBit(bit) }

// Err returns the error which caused Next() to return false.
func (iter *TypeMapIterator) Err() error { return iter.impl.Err() }

// Flush ensures that all pending writes have reached the OS.
func (iter *TypeMapIterator) Flush() error { return iter.impl.Flush() }

// Close flushes writes and frees the resources used by the iterator.
func (iter *TypeMapIterator) Close() error { return iter.impl.Close() }

// Skip is equivalent to calling Next() n times, but faster.
func (iter *TypeMapIterator) Skip(n uint64) bool {
	if n == 1 && iter.primed {
		iter.last = iter.impl.Bit()
	} else if n == 1 {
		iter.last = SType
		iter.primed = true
	} else if n > 1 {
		if !iter.impl.Skip(n - 1) {
			return false
		}
		iter.last = iter.impl.Bit()
		iter.primed = true
		n = 1
	}
	return iter.impl.Skip(n)
}

// BuildTypeMap scans a text, constructing a TypeMap from it.
func BuildTypeMap(text *Text, opts ...Option) (*TypeMap, error) {
	opts = extendOptions(
		opts,
		NumValues(text.Len()+1))

	typeMap, err := NewTypeMap(opts...)
	if err != nil {
		return nil, err
	}

	needClose := true
	defer func() {
		if needClose {
			typeMap.Close()
		}
	}()

	baIter := typeMap.bv.ReverseIterate(0, typeMap.bv.Len())
	defer baIter.Close()

	if !baIter.Next() {
		return nil, baIter.Close()
	}
	baIter.SetBit(SType)

	if text.Len() == 0 {
		if err := baIter.Close(); err != nil {
			return nil, err
		}
		needClose = false
		return typeMap, nil
	}

	if !baIter.Next() {
		return nil, baIter.Close()
	}
	baIter.SetBit(LType)

	if text.Len() == 1 {
		if err := baIter.Close(); err != nil {
			return nil, err
		}
		needClose = false
		return typeMap, nil
	}

	textIter := text.ReverseIterate(0, text.Len())
	defer textIter.Close()

	haveLast := false
	var lastType bool
	var lastSymbol uint64
	for textIter.Next() {
		if !haveLast {
			lastType = LType
			lastSymbol = textIter.Symbol()
			haveLast = true
			continue
		}
		if !baIter.Next() {
			return nil, baIter.Close()
		}
		thisSymbol := textIter.Symbol()
		var bit bool
		if thisSymbol > lastSymbol {
			bit = LType
		} else if thisSymbol == lastSymbol {
			bit = lastType
		} else {
			bit = SType
		}
		baIter.SetBit(bit)
		lastType = bit
		lastSymbol = thisSymbol
	}
	if err := textIter.Close(); err != nil {
		return nil, err
	}
	if err := baIter.Close(); err != nil {
		return nil, err
	}

	needClose = false
	return typeMap, nil
}
