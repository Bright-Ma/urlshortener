package validator

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	v := validator.New()
	v.RegisterValidation("after", func(fl validator.FieldLevel) bool {
		startTime, ok := fl.Field().Interface().(time.Time)
		if !ok {
			return false
		}

		return startTime.After(time.Now())
	})

	return &CustomValidator{
		validator: v,
	}
}

func (c *CustomValidator) Validate(i interface{}) error {
	return c.validator.Struct(i)
}
