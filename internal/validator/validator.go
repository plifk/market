// Package validator can be used to return form validation errors.
package validator

import (
	"errors"
	"fmt"
	"strings"
)

// TemplateErrors returns a FormError with the form errors to be consumed in a template.
// If no errors exist, it returns an empty object.
func TemplateErrors(err error) FormError {
	var fe FormError
	errors.As(err, &fe)
	return fe
}

// FormError for validation.
type FormError []FieldError

// Append form error.
func (f FormError) Append(field string, err ...error) FormError {
	for _, each := range f {
		if each.Field == field {
			each.Add(err...)
			return f
		}
	}
	return append(f, NewFieldError(field, err...))
}

// Get field errors.
func (f FormError) Get(field string) (FieldError, bool) {
	for _, each := range f {
		if each.Field == field {
			return each, true
		}
	}
	return FieldError{}, false
}

// Error message.
func (f FormError) Error() string {
	var errors []string
	for _, each := range f {
		errors = append(errors, fmt.Sprintf("%s field: %v", each.Field, each.Error()))
		if len(errors) == 0 {
			return "<nil>"
		}
	}
	return strings.Join(errors, "; ")
}

// NewFieldError for validation.
func NewFieldError(field string, err ...error) FieldError {
	return FieldError{
		Field:  field,
		Errors: err,
	}
}

// FieldError for validation.
type FieldError struct {
	Field  string
	Errors []error
}

// Add field errors.
func (f FieldError) Add(err ...error) {
	f.Errors = append(f.Errors, err...)
}

// Error message.
func (f FieldError) Error() string {
	var errors []string
	for _, err := range f.Errors {
		if err != nil {
			errors = append(errors, err.Error())
		}

	}
	if len(errors) == 0 {
		return "<nil>"
	}
	return strings.Join(errors, ", ")
}
