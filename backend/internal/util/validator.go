package util

import (
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// InitValidator 注册全局校验器。
func InitValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("noemoji", func(fl validator.FieldLevel) bool {
			return !strings.Contains(fl.Field().String(), "\U0001F600")
		})
	}
}
