package suffixarray

import (
	"bytes"
	"fmt"

	bigarray "github.com/team-spectre/go-bigarray"
)

// Text provides an interface for dealing with very large strings that draw
// symbols from a variable-sized alphabet.
type Text struct {
	ab uint64
	ba bigarray.BigArray
}

// TextIterator iterates over a Text.
//
// The basic usage pattern is:
//
//   iter := text.Iterate(i, j)
//   for iter.Next() {
//     ... // call Index(), Symbol(), and/or SetSymbol()
//   }
//   err := iter.Close()
//   if err != nil {
//     ... // handle error
//   }
//
// Iterators are created in an indeterminate state; the caller must invoke
// Next() to advance to the first item.
//
type TextIterator struct {
	impl bigarray.Iterator
}

// NewText constructs a Text with the given alphabet.
func NewText(alphaSize uint64, opts ...Option) (*Text, error) {
	maxValue := alphaSize
	if maxValue > 1 {
		maxValue--
	}

	opts = extendOptions(
		opts,
		MaxValue(maxValue))

	ba, err := makeBigArray(opts)
	if err != nil {
		return nil, err
	}
	return &Text{ab: alphaSize, ba: ba}, nil
}

// AlphabetSize returns the number of symbols in this text's alphabet.
func (text *Text) AlphabetSize() uint64 { return text.ab }

// Len returns the length of this text in symbols.
func (text *Text) Len() uint64 { return text.ba.Len() }

// SymbolAt returns the symbol at the given index.
func (text *Text) SymbolAt(index uint64) (uint64, error) {
	return text.ba.ValueAt(index)
}

// SetSymbolAt replaces the symbol at the given index.
func (text *Text) SetSymbolAt(index uint64, symbol uint64) error {
	return text.ba.SetValueAt(index, symbol)
}

// Iterate constructs a TextIterator in the forward direction.
func (text *Text) Iterate(i, j uint64) *TextIterator {
	return &TextIterator{impl: text.ba.Iterate(i, j)}
}

// ReverseIterate constructs a TextIterator in the reverse direction.
func (text *Text) ReverseIterate(i, j uint64) *TextIterator {
	return &TextIterator{impl: text.ba.ReverseIterate(i, j)}
}

// ForEach is a convenience method that iterates over the entire TextMap in the
// forward direction.
func (text *Text) ForEach(fn func(uint64, uint64) error) error {
	return bigarray.ForEach(text.ba, fn)
}

// ReverseForEach is a convenience method that iterates over the entire TextMap
// in the reverse direction.
func (text *Text) ReverseForEach(fn func(uint64, uint64) error) error {
	return bigarray.ReverseForEach(text.ba, fn)
}

// CopyFrom copies the symbols from another Text to this one.  The two Texts
// must have the same Len(), and this text must have an AlphabetSize large
// enough to accommodate any symbol in the source text.
func (text *Text) CopyFrom(src *Text) error {
	return text.ba.CopyFrom(src.ba)
}

// Truncate trims the Text to the given length.
func (text *Text) Truncate(length uint64) error { return text.ba.Truncate(length) }

// Freeze makes the Text read-only.
func (text *Text) Freeze() error { return text.ba.Freeze() }

// Flush ensures that all pending writes have reached the OS.
func (text *Text) Flush() error { return text.ba.Flush() }

// Close flushes any writes and frees the resources used by the Text.
func (text *Text) Close() error { return text.ba.Close() }

// Debug returns a human-friendly debugging representation of the Text.
func (text *Text) Debug() string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	err := text.ForEach(func(index uint64, symbol uint64) error {
		if index > 0 {
			buf.WriteByte(' ')
		}
		if text.AlphabetSize() == 256 {
			fmt.Fprintf(&buf, "%q", byte(symbol))
		} else {
			fmt.Fprintf(&buf, "%d", symbol)
		}
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
func (iter *TextIterator) Next() bool { return iter.Skip(1) }

// Skip is equivalent to calling Next() n times, but faster.
func (iter *TextIterator) Skip(n uint64) bool { return iter.impl.Skip(n) }

// Index returns the index of the current symbol.
func (iter *TextIterator) Index() uint64 { return iter.impl.Index() }

// Symbol returns the value of the current symbol.
func (iter *TextIterator) Symbol() uint64 { return iter.impl.Value() }

// SetSymbol replaces the value of the current symbol.
func (iter *TextIterator) SetSymbol(symbol uint64) { iter.impl.SetValue(symbol) }

// Err returns the error which caused Next() to return false.
func (iter *TextIterator) Err() error { return iter.impl.Err() }

// Flush ensures that all pending writes have reached the OS.
func (iter *TextIterator) Flush() error { return iter.impl.Flush() }

// Close flushes writes and frees the resources used by the iterator.
func (iter *TextIterator) Close() error { return iter.impl.Close() }
