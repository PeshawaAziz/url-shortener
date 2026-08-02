package cache

import (
	"context"

	"github.com/bits-and-blooms/bloom/v3"
)

type LocalBloomFilter struct {
	filter *bloom.BloomFilter
}

func NewLocalBloomFilter(m uint, k uint) *LocalBloomFilter {
	return &LocalBloomFilter{
		filter: bloom.New(m, k),
	}
}

func (b *LocalBloomFilter) Exists(ctx context.Context, slug string) bool {
	return b.filter.Test([]byte(slug))
}

func (b *LocalBloomFilter) Add(ctx context.Context, slug string) {
	b.filter.Add([]byte(slug))
}
