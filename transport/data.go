package transport

// DataContainer is a sentinel structure to hold a sole item as Data.
//
// This structure helps APIs exposed to external system clients through transport protocols
// to create a definition of a returning values and thus, a clear separation of different return values
// (i.e. errors).
type DataContainer[T any] struct {
	// Data the item to be tagged as data value.
	Data T `json:"data"`
}

// PageResponse is a structure to hold a paginated response.
//
// This structure helps APIs exposed to external system clients through transport protocols
// to create a definition of a paginated response and thus, a clear separation of different return values.
type PageResponse[T any] struct {
	TotalItems        int    `json:"total_items"`
	PreviousPageToken string `json:"previous_page_token"`
	NextPageToken     string `json:"next_page_token"`
	Items             []T    `json:"items"`
}
