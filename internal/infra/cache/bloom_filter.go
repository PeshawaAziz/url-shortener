package cache

import (
	"context"

	"github.com/bits-and-blooms/bloom/v3"
)

type BloomFilterAdapter struct {
	filter *bloom.BloomFilter
}

func NewBloomFilterAdapter(m uint, k uint) *BloomFilterAdapter {
	return &BloomFilterAdapter{
		filter: bloom.New(m, k), // m = number of bits, k = number of hash functions.
	}
}

func (b *BloomFilterAdapter) Add(ctx context.Context, slug string) {
	b.filter.Add([]byte(slug))
}

func (b *BloomFilterAdapter) Exists(ctx context.Context, slug string) bool {
	return b.filter.Test([]byte(slug))
}

func (b *BloomFilterAdapter) LoadExisting(slugs []string) {
	for _, s := range slugs {
		b.filter.Add([]byte(s))
	}
}
