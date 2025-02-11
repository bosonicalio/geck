package criteria

import (
	"errors"

	"github.com/samber/lo"

	"github.com/hadroncorp/geck/persistence"
)

// TranslateFields translates the fields from `v` ([Criteria]) using `t` as mapper component.
//
// Leaves the field unmodified if `t` has no mapping assigned.
func TranslateFields(t *persistence.FieldTranslator, v *Criteria) error {
	if t == nil || v == nil {
		return errors.New("cannot translate fields, invalid parameters")
	}
	v.Sorting.Field = lo.CoalesceOrEmpty(t.Translate(v.Sorting.Field), v.Sorting.Field)
	for i := range v.Filters {
		v.Filters[i].Field = lo.CoalesceOrEmpty(t.Translate(v.Filters[i].Field), v.Filters[i].Field)
	}
	return nil
}
