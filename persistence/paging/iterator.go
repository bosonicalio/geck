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

package paging

import (
	"io"
)

// - Iterator -

const _defaultPageSize = 100

// Iterator is a generic iterator for paginated data.
// It fetches data using a provided [FetchFunc] and allows iteration over the items.
// The iterator supports both forward and reverse pagination, controlled by the [WithIteratorReverse] option.
//
// Usage example:
// ```go
// ctx := context.Background()
//
//	fetchFunc := func(opts ...paging.Option) (*paging.Page[YourType], error) {
//	    // Implement your fetching logic here, e.g., querying a database or an API.
//	    return yourPage, nil
//	}
//
// iterator := paging.NewIterator(ctx, fetchFunc,
//
//	paging.WithIteratorPageSize(50), // Optional: set custom page size
//	paging.WithIteratorReverse(false), // Optional: set reverse pagination
//	paging.WithIteratorPageToken("your-page-token"), // Optional: set initial page token
//
// )
//
//	for iterator.HasNext() {
//	    item, err := iterator.Next()
//	    if err != nil {
//	        if err == io.EOF {
//	            break // No more items to iterate
//	        }
//	        // Handle other errors
//	        panic(err)
//	    }
//	    // Process the item
//	    fmt.Println(item)
//	}
//
// The iterator will automatically handle pagination, fetching new pages as needed.
// The items are returned in the order they are fetched, and the iterator will stop when there are no more items to fetch.
type Iterator[T any] struct {
	fetchFunc    FetchFunc[T]
	currentIndex int
	items        []T

	isReverse     bool
	pageSize      int
	lastPageToken string
}

// FetchFunc is a function type that defines how to fetch a page of items.
type FetchFunc[T any] func(opts ...Option) (*Page[T], error)

// NewIterator creates a new Iterator instance with the provided fetch function and options.
func NewIterator[T any](fetchFunc FetchFunc[T], opts ...IteratorOption) *Iterator[T] {
	options := &iteratorOptions{}
	for _, opt := range opts {
		opt(options)
	}
	options.setDefaults()
	return &Iterator[T]{
		fetchFunc:     fetchFunc,
		isReverse:     options.isReverse,
		pageSize:      options.pageSize,
		lastPageToken: options.pageToken,
	}
}

// HasNext checks if there are more items to iterate over.
func (i *Iterator[T]) HasNext() bool {
	return i.items == nil || i.currentIndex < len(i.items) || i.hasNextPage()
}

// Next retrieves the next item from the iterator.
func (i *Iterator[T]) Next() (T, error) {
	if i.currentIndex >= len(i.items) {
		if err := i.loadNextPage(); err != nil {
			var zero T
			return zero, err
		}
	}
	item := i.items[i.currentIndex]
	i.currentIndex++
	return item, nil
}

func (i *Iterator[T]) hasNextPage() bool {
	return i.lastPageToken != ""
}

func (i *Iterator[T]) loadNextPage() error {
	if i.items != nil && !i.hasNextPage() {
		return io.EOF
	}

	opts := make([]Option, 0, 2)
	if i.lastPageToken != "" {
		opts = append(opts, WithPageToken(i.lastPageToken))
	}
	opts = append(opts, WithLimit(i.pageSize))
	page, err := i.fetchFunc(opts...)
	if err != nil {
		return err
	} else if page == nil || len(page.Items) == 0 {
		return io.EOF
	}
	if i.items == nil {
		i.items = make([]T, 0, i.pageSize) // allocate result buffer once on first fetch op to reuse it later
	} else {
		i.items = i.items[:0] // Reset the result buffer to reuse it, but keep the capacity to avoid reallocating
	}

	if i.isReverse {
		i.lastPageToken = page.PreviousPageToken
	} else {
		i.lastPageToken = page.NextPageToken
	}
	i.items = append(i.items, page.Items...)
	i.currentIndex = 0 // Reset the current index to the start of the new items
	return nil
}

// -- Options --

type iteratorOptions struct {
	pageSize  int
	pageToken string
	isReverse bool
}

func (i *iteratorOptions) setDefaults() {
	if i.pageSize <= 0 {
		i.pageSize = _defaultPageSize
	}
}

type IteratorOption func(*iteratorOptions)

// WithIteratorPageSize sets the page size for the iterator.
func WithIteratorPageSize(size int) IteratorOption {
	return func(opts *iteratorOptions) {
		opts.pageSize = size
	}
}

// WithIteratorReverse sets whether the iterator should fetch items in reverse order.
func WithIteratorReverse(isReverse bool) IteratorOption {
	return func(opts *iteratorOptions) {
		opts.isReverse = isReverse
	}
}

// WithIteratorPageToken sets the initial page token for the iterator.
func WithIteratorPageToken(token string) IteratorOption {
	return func(opts *iteratorOptions) {
		opts.pageToken = token
	}
}
