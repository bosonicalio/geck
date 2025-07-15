package blob

// Bucket is a component providing functionality for uploading and removing objects.
type Bucket interface {
	ObjectUploader
	ObjectRemover
}
