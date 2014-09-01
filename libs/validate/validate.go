package validate

import (
	"github.com/asaskevich/govalidator"
	"github.com/pengux/check"
)

//reexports
type (
	NonEmpty  check.NonEmpty
	MinChar   check.MinChar
	MaxChar   check.MaxChar
	Email     check.Email
	Composite check.Composite
	Validator check.Validator
	Struct    check.Struct
)

func (validator NonEmpty) Validate(v interface{}) check.Error {
	return check.NonEmpty{}.Validate(v)
}

func (validator MinChar) Validate(v interface{}) check.Error {
	return check.MinChar{validator.Constraint}.Validate(v)
}

func (validator MaxChar) Validate(v interface{}) check.Error {
	return check.MaxChar{validator.Constraint}.Validate(v)
}

func (validator Email) Validate(v interface{}) check.Error {
	return check.Email{}.Validate(v)
}

func (validator Composite) Validate(v interface{}) check.Error {
	return check.Composite(validator).Validate(v)
}

func (validator Struct) Validate(v interface{}) check.StructError {
	return check.Struct(validator).Validate(v)
}

func init() {
	check.ErrorMessages["alnum"] = "only letters and numbers are allowed."
}

type (
	Alphanumeric struct{}
)

func (validator Alphanumeric) Validate(v interface{}) check.Error {
	if !govalidator.IsAlphanumeric(v.(string)) {
		return check.NewValidationError("alnum")
	}

	return nil
}
