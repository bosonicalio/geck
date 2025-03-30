package validation

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"

	"github.com/hadroncorp/geck/syserr"
)

// Validator is a utility component used by systems to validate a certain value and/or structure.
type Validator interface {
	// Validate validates the given value.
	Validate(ctx context.Context, v any) error
}

// GoPlaygroundValidator a concrete implementation of Validator using go-playground/validator package.
type GoPlaygroundValidator struct {
	driver   StructFieldDriver
	validate *validator.Validate
}

// compile-time assertion
var _ Validator = (*GoPlaygroundValidator)(nil)

// NewGoPlaygroundValidator allocates a new GoPlaygroundValidator instance.
func NewGoPlaygroundValidator(config ValidatorConfig) GoPlaygroundValidator {
	v := validator.New()
	_ = v.RegisterValidation("date", validateDate)
	return GoPlaygroundValidator{
		driver:   config.StructFieldDriver,
		validate: v,
	}
}

func validateDate(fl validator.FieldLevel) bool {
	val, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	_, err := time.Parse(time.DateOnly, val)
	return err == nil
}

// Validate validates the given value. Returns error if one or more validations failed.
func (g GoPlaygroundValidator) Validate(ctx context.Context, v any) error {
	rawErr := g.validate.StructCtx(ctx, v)
	if rawErr == nil {
		return nil
	}

	var errsValidation validator.ValidationErrors
	ok := errors.As(rawErr, &errsValidation)
	if !ok {
		return rawErr
	}

	typeof := reflect.TypeOf(v)
	customStructFields := g.newFieldTagMap(typeof, g.driver.String(), typeof.Name(), "")
	errs := make([]error, 0, len(errsValidation))
	for _, errValidation := range errsValidation {
		// use codec field name if any
		field := lo.CoalesceOrEmpty(customStructFields[errValidation.Namespace()],
			strings.TrimPrefix(errValidation.Namespace(), typeof.Name()+"."))
		switch errValidation.Tag() {
		case "required":
			errs = append(errs, syserr.NewMissingValue(field))
		case "oneof":
			values := strings.Split(errValidation.Param(), " ")
			errs = append(errs, syserr.NewNotOneOf(field, values...))
		case "eq", "eq_ignore_case":
			errs = append(errs, syserr.NewNotEquals(field, errValidation.Param()))
		case "ne", "ne_ignore_case":
			errs = append(errs, syserr.NewEquals(field, errValidation.Param()))
		case "len":
			expLen, _ := strconv.Atoi(errValidation.Param())
			errs = append(errs, syserr.NewInvalidLength(field, expLen))
		case "min", "gt", "gte":
			minVal, _ := strconv.Atoi(errValidation.Param())
			if errValidation.Tag() == "gt" {
				minVal = minVal + 1
			}
			errs = append(errs, syserr.NewBelowLimit(field, minVal))
		case "max", "lt", "lte":
			maxVal, _ := strconv.Atoi(errValidation.Param())
			if errValidation.Tag() == "lt" {
				maxVal = maxVal - 1
			}
			errs = append(errs, syserr.NewAboveLimit(field, maxVal))
		default:
			errs = append(errs, syserr.NewInvalidFormat(field, errValidation.Tag()))
		}
	}
	return errors.Join(errs...)
}

func (g GoPlaygroundValidator) newFieldTagMap(typeof reflect.Type, tag, prefixField, prefixTag string) map[string]string {
	tagMap := make(map[string]string, typeof.NumField())
	for i := 0; i < typeof.NumField(); i++ {
		field := typeof.Field(i)
		structTag := field.Tag.Get(tag)
		if structTag == "" {
			continue
		}

		fieldName := field.Name
		if prefixField != "" {
			fieldName = prefixField + "." + fieldName
		}
		if prefixTag != "" {
			structTag = prefixTag + "." + structTag
		}
		tagMap[fieldName] = structTag
		if field.Type.Kind() == reflect.Struct {
			tagMap = lo.Assign(tagMap, g.newFieldTagMap(field.Type, tag, fieldName, structTag))
		}
	}
	return tagMap
}
