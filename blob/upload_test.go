package blob_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/tesserical/geck/blob"
	"github.com/tesserical/geck/blobmock"
)

func TestUploadAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	uploader := blobmock.NewMockFileUploader(ctrl)
	uploader.EXPECT().
		Upload(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(4).
		Return(error(nil))
	err := blob.UploadAll(t.Context(), uploader,
		blob.WithBatchUploaderMaxProcs(2),
		blob.WithBatchUploadItemBytes("test-key-0", []byte("test-data-0")),
		blob.WithBatchUploadItem("test-key-1", bytes.NewReader([]byte("test-data-1"))),
		blob.WithBatchUploadItemBytes("test-key-2", []byte("test-data-2")),
		blob.WithBatchUploadItem("test-key-3", bytes.NewReader([]byte("test-data-3"))),
	)
	assert.NoError(t, err)
}

func TestUploadAllFromFS(t *testing.T) {
	ctrl := gomock.NewController(t)
	uploader := blobmock.NewMockFileUploader(ctrl)
	uploader.EXPECT().
		Upload(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(2).
		Return(error(nil))
	err := blob.UploadAllFromFS(t.Context(), uploader, os.DirFS("testdata"),
		blob.WithBatchUploadFSMaxProcs(2),
	)
	assert.NoError(t, err)
}
