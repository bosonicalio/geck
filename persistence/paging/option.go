package paging

// Options represents the options for pagination.
type Options struct {
	limit     int
	pageToken string
}

// Option represents an option for pagination.
type Option func(*Options)

// WithLimit sets the maximum number of items to return.
func WithLimit(limit int) Option {
	return func(o *Options) {
		o.limit = limit
	}
}

// WithPageToken sets the page token to use for pagination.
func WithPageToken(token string) Option {
	return func(o *Options) {
		o.pageToken = token
	}
}

// Limit returns the limit option.
func (o Options) Limit() int {
	return o.limit
}

// PageToken returns the page token option.
func (o Options) PageToken() string {
	return o.pageToken
}

// HasPageToken returns true if the page token is set.
func (o Options) HasPageToken() bool {
	return o.pageToken != ""
}
