package transport

import (
	"encoding"
	"fmt"
)

// MimeType represents a MIME type.
type MimeType uint16

const (
	MimeTypeUnknown = MimeType(iota)
	MimeTypeJSON
	MimeTypeXML
	MimeTypeYAML
	MimeTypeProtobuf
	MimeTypeAvro
	MimeTypeThrift
	MimeTypeMsgPack
	MimeTypeForm
	MimeTypeMultipartForm
	MimeTypeOctetStream
	MimeTypeText
	MimeTypeHTML
	MimeTypeCSV
	MimeTypePNG
	MimeTypeJPEG
	MimeTypeGIF
	MimeTypeSVG
	MimeTypeBMP
	MimeTypeTIFF
	MimeTypeWebP
	MimeTypeHEIF
	MimeTypeICO
	MimeTypeAVIF
	MimeTypeWebM
	MimeTypeMP4
	MimeTypeMKV
	MimeTypeAVI
	MimeTypeWMV
	MimeTypeFLV
	MimeTypeMOV
	MimeTypeMPEG
	MimeTypeMP3
	MimeTypeWAV
)

// compile-time assertion
var (
	_ fmt.Stringer             = MimeTypeUnknown
	_ encoding.TextUnmarshaler = (*MimeType)(nil)
	_ encoding.TextMarshaler   = (*MimeType)(nil)
)

var (
	_mimeTypeMap = map[string]MimeType{
		"application/json":                  MimeTypeJSON,
		"application/xml":                   MimeTypeXML,
		"application/x-yaml":                MimeTypeYAML,
		"application/x-protobuf":            MimeTypeProtobuf,
		"application/avro":                  MimeTypeAvro,
		"application/x-thrift":              MimeTypeThrift,
		"application/msgpack":               MimeTypeMsgPack,
		"application/x-www-form-urlencoded": MimeTypeForm,
		"multipart/form-data":               MimeTypeMultipartForm,
		"application/octet-stream":          MimeTypeOctetStream,
		"text/plain":                        MimeTypeText,
		"text/html":                         MimeTypeHTML,
		"text/csv":                          MimeTypeCSV,
		"image/png":                         MimeTypePNG,
		"image/jpeg":                        MimeTypeJPEG,
		"image/gif":                         MimeTypeGIF,
		"image/svg+xml":                     MimeTypeSVG,
		"image/bmp":                         MimeTypeBMP,
		"image/tiff":                        MimeTypeTIFF,
		"image/webp":                        MimeTypeWebP,
		"image/heif":                        MimeTypeHEIF,
		"image/x-icon":                      MimeTypeICO,
		"image/avif":                        MimeTypeAVIF,
		"video/webm":                        MimeTypeWebM,
		"video/mp4":                         MimeTypeMP4,
		"video/x-matroska":                  MimeTypeMKV,
		"video/x-msvideo":                   MimeTypeAVI,
		"video/x-ms-wmv":                    MimeTypeWMV,
		"video/x-flv":                       MimeTypeFLV,
		"video/quicktime":                   MimeTypeMOV,
		"video/mpeg":                        MimeTypeMPEG,
		"audio/mpeg":                        MimeTypeMP3,
		"audio/wav":                         MimeTypeWAV,
	}
	_mimeTypeStringMap = map[MimeType]string{
		MimeTypeUnknown:       "application/octet-stream",
		MimeTypeJSON:          "application/json",
		MimeTypeXML:           "application/xml",
		MimeTypeYAML:          "application/x-yaml",
		MimeTypeProtobuf:      "application/x-protobuf",
		MimeTypeAvro:          "application/avro",
		MimeTypeThrift:        "application/x-thrift",
		MimeTypeMsgPack:       "application/msgpack",
		MimeTypeForm:          "application/x-www-form-urlencoded",
		MimeTypeMultipartForm: "multipart/form-data",
		MimeTypeOctetStream:   "application/octet-stream",
		MimeTypeText:          "text/plain",
		MimeTypeHTML:          "text/html",
		MimeTypeCSV:           "text/csv",
		MimeTypePNG:           "image/png",
		MimeTypeJPEG:          "image/jpeg",
		MimeTypeGIF:           "image/gif",
		MimeTypeSVG:           "image/svg+xml",
		MimeTypeBMP:           "image/bmp",
		MimeTypeTIFF:          "image/tiff",
		MimeTypeWebP:          "image/webp",
		MimeTypeHEIF:          "image/heif",
		MimeTypeICO:           "image/x-icon",
		MimeTypeAVIF:          "image/avif",
		MimeTypeWebM:          "video/webm",
		MimeTypeMP4:           "video/mp4",
		MimeTypeMKV:           "video/x-matroska",
		MimeTypeAVI:           "video/x-msvideo",
		MimeTypeWMV:           "video/x-ms-wmv",
		MimeTypeFLV:           "video/x-flv",
		MimeTypeMOV:           "video/quicktime",
		MimeTypeMPEG:          "video/mpeg",
		MimeTypeMP3:           "audio/mpeg",
		MimeTypeWAV:           "audio/wav",
	}
)

// NewMimeType creates a new [MimeType] from a string value.
func NewMimeType(s string) MimeType {
	return _mimeTypeMap[s]
}

// String returns the string representation of the MIME type.
func (m MimeType) String() string {
	return _mimeTypeStringMap[m]
}

// MarshalText marshals the MIME type into text.
func (m MimeType) MarshalText() (text []byte, err error) {
	return []byte(m.String()), nil
}

// UnmarshalText unmarshals the MIME type from text.
func (m *MimeType) UnmarshalText(text []byte) error {
	*m = _mimeTypeMap[string(text)]
	return nil
}
