package blob

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"runtime"
	"sync"
)

// - Batch Uploader -

// UploadAll uploads multiple items concurrently to the blob storage using the provided ObjectUploader.
func UploadAll(ctx context.Context, uploader ObjectUploader, opts ...BatchUploaderOption) error {
	config := newBatchUploaderOpts()
	for _, opt := range opts {
		opt(config)
	}

	sem := make(chan struct{}, config.maxProcs) // Limit concurrency to 10 uploads
	errs := make([]error, 0, len(config.items))
	errMu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	wg.Add(len(config.items))
	for i := range config.items {
		sem <- struct{}{} // Acquire a slot
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }() // Release the slot when done
			if err := uploader.Upload(ctx, config.items[i].key, config.items[i].data); err != nil {
				errMu.Lock()
				errs = append(errs, err)
				errMu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	return errors.Join(errs...)
}

type uploadItem struct {
	key  string
	data io.Reader
}

// -- Options --

type batchUploaderOpts struct {
	maxProcs int
	items    []uploadItem
}

func newBatchUploaderOpts() *batchUploaderOpts {
	return &batchUploaderOpts{
		maxProcs: min(10, runtime.NumCPU()), // Default to 10 or number of CPU cores
	}
}

type BatchUploaderOption func(*batchUploaderOpts)

// WithBatchUploaderMaxProcs sets the maximum number of concurrent uploads.
func WithBatchUploaderMaxProcs(maxProcs int) BatchUploaderOption {
	return func(opts *batchUploaderOpts) {
		if maxProcs > 0 {
			opts.maxProcs = maxProcs
		}
	}
}

// WithBatchUploadItem sets the items to be uploaded.
func WithBatchUploadItem(key string, data io.Reader) BatchUploaderOption {
	return func(opts *batchUploaderOpts) {
		if key == "" || data == nil {
			return
		}
		if opts.items == nil {
			opts.items = make([]uploadItem, 0, 1)
		}
		opts.items = append(opts.items, uploadItem{
			key:  key,
			data: data,
		})
	}
}

// WithBatchUploadItemBytes sets a byte slice as an item to be uploaded.
func WithBatchUploadItemBytes(key string, data []byte) BatchUploaderOption {
	return func(opts *batchUploaderOpts) {
		if key == "" || data == nil {
			return
		}
		if opts.items == nil {
			opts.items = make([]uploadItem, 0, 1)
		}
		opts.items = append(opts.items, uploadItem{
			key:  key,
			data: bytes.NewReader(data),
		})
	}
}

// - Filesystem Uploader -

// UploadAllFromFS uploads all files from the provided filesystem to the blob storage using the given uploader.
func UploadAllFromFS(ctx context.Context, uploader ObjectUploader, fsys fs.FS, opts ...BatchUploaderFSOption) error {
	if fsys == nil {
		return errors.New("filesystem is nil")
	}
	config := &batchUploaderFSOpts{}
	for _, opt := range opts {
		opt(config)
	}
	items := make([]uploadItem, 0, 8) // Preallocate space for items to be uploaded
	// Iterate through the files in the provided filesystem
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil // Skip directories
		}

		file, err := fsys.Open(path)
		if err != nil {
			return err
		}

		items = append(items, uploadItem{
			key:  path,
			data: file,
		})
		return nil
	})
	if err != nil {
		return err
	} else if len(items) == 0 {
		return errors.New("no files found to upload")
	}
	defer func() {
		for i := range items {
			closer, ok := items[i].data.(io.Closer)
			if !ok {
				continue // Skip if the item is not closable
			}
			errClose := closer.Close()
			if errClose != nil {
				err = errors.Join(err, errClose)
			}
		}
	}()
	baseOpts := make([]BatchUploaderOption, 0, len(items)+1)
	baseOpts = append(baseOpts, WithBatchUploaderMaxProcs(config.baseOpts.maxProcs))
	for i := range items {
		baseOpts = append(baseOpts, WithBatchUploadItem(items[i].key, items[i].data))
	}

	return UploadAll(ctx, uploader, baseOpts...)
}

// -- Options --

// Defining a separate options structure for filesystem uploads, ensuring it can be extended independently.

type batchUploaderFSOpts struct {
	baseOpts batchUploaderOpts
}

// BatchUploaderFSOption defines a function type for configuring the batch uploader options.
type BatchUploaderFSOption func(*batchUploaderFSOpts)

// WithBatchUploadFSMaxProcs sets the maximum number of concurrent uploads for filesystem uploads.
func WithBatchUploadFSMaxProcs(maxProcs int) BatchUploaderFSOption {
	return func(opts *batchUploaderFSOpts) {
		if maxProcs > 0 {
			opts.baseOpts.maxProcs = maxProcs
		}
	}
}
