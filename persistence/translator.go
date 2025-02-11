package persistence

// FieldTranslator is a utility component for persistence systems to create field mappings between A and B.
//
// Useful for use cases like criteria.Criteria processing where fields at application layer are different
// from the fields declared at the persistence layer.
type FieldTranslator struct {
	Source map[string]string
}

// Translate retrieves the mapped value for `field`.
func (t FieldTranslator) Translate(field string) string {
	if t.Source == nil {
		return ""
	}
	return t.Source[field]
}
