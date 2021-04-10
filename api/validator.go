package api

import (
	"github.com/Jiaget/simplebank/util"
	"github.com/go-playground/validator/v10"
)

var validateCurrency validator.Func = func(fl validator.FieldLevel) bool {
	if currency, ok := fl.Field().Interface().(string); ok {
		return util.IsSupporttedCurrency(currency)
	}
	return false
}
