package paging

// Page is a segment of data retrieved from a source (i.e. persistence stores).
//
// As in books, sources using [Page] may allow callers to retrieve a certain part of their whole data set.
// This also allows callers to fetch previous or following results easily.
type Page[T any] struct {
	TotalItems        int
	PreviousPageToken string
	NextPageToken     string
	Items             []T
}
