package handler

import (
	"errors"
	"fmt"
	"invest/model"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	Error       bool
	FailedField string
	Tag         string
	Value       any
}

var myValidator = validator.New()

func init() {
	myValidator.RegisterValidation("date", func(fl validator.FieldLevel) bool {

		isDate, err := regexp.MatchString(fl.Field().String(), `^(\d{4}-\d{2}-\d{2})?$`)
		if err != nil {
			return false
		}
		return isDate
	})

	myValidator.RegisterValidation("market_status", func(fl validator.FieldLevel) bool {
		return fl.Field().Uint() >= 1 && fl.Field().Uint() <= 5
	})

	myValidator.RegisterValidation("category", func(fl validator.FieldLevel) bool {
		return fl.Field().Uint() >= 1 && fl.Field().Uint() <= model.CategoryLength()
	})
}

func validCheck(s any) error {
	if errs := validate(s); len(errs) > 0 && errs[0].Error {
		errMsgs := make([]string, 0)
		for _, err := range errs {
			if err.Tag == "required" {
				errMsgs = append(errMsgs, fmt.Sprintf(
					"필수 필드 %s 누락",
					err.FailedField,
				))
			} else {
				errMsgs = append(errMsgs, fmt.Sprintf(
					"%s 타입의 필드 %s가 올바르지 않음",
					err.Tag,
					err.FailedField,
				))
			}
		}
		return errors.New(strings.Join(errMsgs, ", "))
	}
	return nil
}

func validate(data any) []ErrorResponse {
	validationErrors := []ErrorResponse{}
	errs := myValidator.Struct(data)
	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			var elem ErrorResponse

			elem.FailedField = err.Field()
			elem.Tag = err.Tag()
			elem.Value = err.Value()
			elem.Error = true

			validationErrors = append(validationErrors, elem)
		}
	}
	return validationErrors
}

func dateCheck(s string) bool {
	isDate, err := regexp.MatchString(`^(\d{4}-\d{2}-\d{2})?$`, s)
	if err != nil {
		return false
	}
	return isDate
}
