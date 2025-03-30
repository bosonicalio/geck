package structs

import (
	"reflect"
	"strings"

	"github.com/samber/lo"
)

type structValueOptions struct {
	tag          string
	tagSeparator string
}

// StructValueOption is a routine used to specify configurations to [GetStructValue].
type StructValueOption func(*structValueOptions)

// WithTag indicates the tag to use as fallback during [GetStructValue] execution.
//
// [GetStructValue] will split the tag by `,` (or the one provided if used [WithTagSeparator]) if it
// defines multi-values (e.g. `foo:"bar,baz"`) and will use the first value provided.
func WithTag(t string) StructValueOption {
	return func(o *structValueOptions) {
		o.tag = t
	}
}

// WithTagSeparator indicates the tag separator pattern to split tag (provided by [WithTag]).
//
// `,` will be used as default if no pattern provided.
func WithTagSeparator(s string) StructValueOption {
	return func(o *structValueOptions) {
		o.tagSeparator = s
	}
}

// GetStructValue retrieves a value from `v`.
// The type of `v` MUST be a struct, otherwise returns nil.
//
// Use [StructValueOption] derivatives (e.g. [WithTag], [WithTagSeparator]) to customize how this routine handles
// edge cases.
func GetStructValue(v any, fieldName string, opts ...StructValueOption) any {
	options := structValueOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	typeOf := reflect.TypeOf(v)
	valueOf := reflect.ValueOf(v)
	if typeOf.Kind() == reflect.Ptr {
		typeOf = typeOf.Elem()
		valueOf = reflect.New(typeOf.Elem())
	}

	if typeOf.Kind() != reflect.Struct {
		return nil
	}

	val := valueOf.FieldByName(fieldName)
	if val != (reflect.Value{}) {
		return val.Interface()
	}

	if options.tag == "" {
		return nil
	}

	options.tagSeparator = lo.CoalesceOrEmpty(options.tagSeparator, ",")
	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)
		alias, ok := field.Tag.Lookup(options.tag)
		aliasSplit := strings.Split(alias, options.tagSeparator)
		if ok && fieldName == aliasSplit[0] {
			return valueOf.Field(i).Interface()
		}
	}
	return nil
}
