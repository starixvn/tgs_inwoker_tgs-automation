// helper/bloom_filter.go
package helper

import (
	"github.com/bits-and-blooms/bloom/v3"
)

// BloomFilter is a wrapper around the bits-and-blooms BloomFilter
type BloomFilter struct {
	filter *bloom.BloomFilter
}

// NewBloomFilter creates a new BloomFilter with estimated parameters
func NewBloomFilter() *BloomFilter {
	n := uint(20000)    // Estimated number of items
	fp := float64(0.01) // Desired false positive rate (1%)
	filter := bloom.NewWithEstimates(n, fp)
	return &BloomFilter{
		filter: filter,
	}
}

// Contains checks if the item is possibly in the Bloom filter
func (bf *BloomFilter) Contains(item string) bool {
	return bf.filter.Test([]byte(item))
}

// Add inserts the item into the Bloom filter
func (bf *BloomFilter) Add(item string) {
	bf.filter.Add([]byte(item))
}
