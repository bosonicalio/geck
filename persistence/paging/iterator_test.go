/*
 * Copyright (c) 2025 Tesserical s.r.l. All rights reserved.
 *
 * This source code is the property of Tesserical s.r.l. and is intended for internal use only.
 * Unauthorized copying, distribution, or disclosure of this code, in whole or in part, is strictly prohibited.
 *
 * For internal development purposes only. Not for public release.
 *
 * For inquiries, contact: legal@tesserical.com
 *
 */

package paging_test

import (
	"io"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tesserical/geck/persistence/paging"
)

func TestIterator_Multipage(t *testing.T) {
	iterCount := 0
	iterator := paging.NewIterator(
		func(opts ...paging.Option) (*paging.Page[string], error) {
			if iterCount == 2 {
				return nil, nil
			}
			defer func() {
				iterCount++
			}()
			return &paging.Page[string]{
				Items:         []string{strconv.Itoa(iterCount)},
				NextPageToken: "next-page-token",
			}, nil
		},
		paging.WithIteratorPageSize(1), // Set page size to 1 for testing
	)
	require.NotNil(t, iterator)

	// Fetch first page
	assert.True(t, iterator.HasNext())
	item, err := iterator.Next()
	assert.NoError(t, err)
	assert.Equal(t, "0", item)

	// Fetch second page
	assert.True(t, iterator.HasNext())
	item, err = iterator.Next()
	assert.NoError(t, err)
	assert.Equal(t, "1", item)

	// Try to fetch third page (will fail as provider has no more pages)
	assert.True(t, iterator.HasNext())
	item, err = iterator.Next()
	assert.ErrorIs(t, err, io.EOF)
	assert.Empty(t, item)
}

func TestIterator_Single_Page(t *testing.T) {
	iterCount := 0
	iterator := paging.NewIterator(
		func(opts ...paging.Option) (*paging.Page[string], error) {
			if iterCount == 1 {
				return nil, nil
			}
			defer func() {
				iterCount++
			}()
			return &paging.Page[string]{
				Items:         []string{strconv.Itoa(iterCount)},
				NextPageToken: "next-page-token",
			}, nil
		},
		paging.WithIteratorPageSize(1), // Set page size to 1 for testing
	)
	require.NotNil(t, iterator)

	// Fetch first page
	assert.True(t, iterator.HasNext())
	item, err := iterator.Next()
	assert.NoError(t, err)
	assert.Equal(t, "0", item)

	// Try to fetch third page (will fail as provider has no more pages)
	assert.True(t, iterator.HasNext())
	item, err = iterator.Next()
	assert.ErrorIs(t, err, io.EOF)
	assert.Empty(t, item)
}
