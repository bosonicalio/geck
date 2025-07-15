package validation

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"

	"github.com/tesserical/geck/syserr"
)

// Validator is a utility component used by systems to validate structures.
type Validator interface {
	// Validate validates the given structure (v).
	Validate(ctx context.Context, v any) error
}

// -- Options --

type options struct {
	codecDriver CodecDriver
	customRules map[string]ValidateFunc
}

func newOptions(opts ...Option) *options {
	config := &options{
		codecDriver: JSONDriver,
	}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

// Option is a function that modifies the validator behavior.
type Option func(*options)

// WithCodecDriver sets the codec driver to be used by the validator.
func WithCodecDriver(driver CodecDriver) Option {
	return func(o *options) {
		o.codecDriver = driver
	}
}

// WithRules adds a set of custom validation rules to the validator.
func WithRules(rules ...Rule) Option {
	return func(o *options) {
		if o.customRules == nil {
			o.customRules = make(map[string]ValidateFunc, len(rules))
		}
		for i := range rules {
			o.customRules[rules[i].Name] = rules[i].ValidateFunc
		}
	}
}

// WithRuleFunc adds a custom validation rule routine to the validator.
func WithRuleFunc(name string, fn ValidateFunc) Option {
	return func(o *options) {
		if o.customRules == nil {
			o.customRules = make(map[string]ValidateFunc, 1)
		}
		o.customRules[name] = fn
	}
}

// -- Go Playground Validator --

// GoPlaygroundValidator a concrete implementation of Validator using go-playground/validator package.
type GoPlaygroundValidator struct {
	driver   CodecDriver
	validate *validator.Validate
}

// compile-time assertion
var _ Validator = (*GoPlaygroundValidator)(nil)

// NewGoPlaygroundValidator allocates a new GoPlaygroundValidator instance.
func NewGoPlaygroundValidator(opts ...Option) GoPlaygroundValidator {
	config := newOptions(opts...)
	v := validator.New()
	for name, fn := range config.customRules {
		_ = v.RegisterValidation(name, func(fl validator.FieldLevel) bool {
			return fn(name, fl.Field().Interface())
		})
	}
	return GoPlaygroundValidator{
		driver:   config.codecDriver,
		validate: v,
	}
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
