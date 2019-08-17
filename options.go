package suffixarray

import (
	"io"
	"sync"

	bigarray "github.com/team-spectre/go-bigarray"
	bigbitvector "github.com/team-spectre/go-bigbitvector"
)

type Option struct {
	BigArrayOption     bigarray.Option
	BigBitVectorOption bigbitvector.Option
}

func NumValues(size uint64) Option {
	return Option{
		bigarray.NumValues(size),
		bigbitvector.NumValues(size),
	}
}

func MaxValue(max uint64) Option {
	return Option{
		bigarray.MaxValue(max),
		nil,
	}
}

func BytesPerValue(bpv uint8) Option {
	return Option{
		bigarray.BytesPerValue(bpv),
		nil,
	}
}

func OnDiskThreshold(size uint64) Option {
	return Option{
		bigarray.OnDiskThreshold(size),
		bigbitvector.OnDiskThreshold(size),
	}
}

func PageSize(size uint) Option {
	return Option{
		bigarray.PageSize(size),
		bigbitvector.PageSize(size),
	}
}

func WithPool(pool *sync.Pool) Option {
	return Option{
		bigarray.WithPool(pool),
		bigbitvector.WithPool(pool),
	}
}

func WithFile(file bigarray.File) Option {
	return Option{
		bigarray.WithFile(file),
		bigbitvector.WithFile(file),
	}
}

func WithReadOnlyFile(file io.ReaderAt) Option {
	return Option{
		bigarray.WithReadOnlyFile(file),
		bigbitvector.WithReadOnlyFile(file),
	}
}
