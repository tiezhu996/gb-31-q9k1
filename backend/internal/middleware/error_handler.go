package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/util"
)

// ErrorHandler 统一错误处理：service 返回的 AppError / 哨兵错误转为 JSON。
func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		var appErr *util.AppError
		if errors.As(err, &appErr) {
			util.Fail(c, appErr.HTTP, appErr.Code, appErr.Message)
			return
		}
		var verr validator.ValidationErrors
		if errors.As(err, &verr) {
			util.Fail(c, http.StatusUnprocessableEntity, constants.CodeValidationFailed, constants.MsgInvalidParam)
			return
		}
		switch {
		case errors.Is(err, constants.ErrNotFound):
			util.Fail(c, http.StatusNotFound, constants.CodeNotFound, constants.MsgNotFound)
		case errors.Is(err, constants.ErrUnauthorized):
			util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
		case errors.Is(err, constants.ErrForbidden):
			util.Fail(c, http.StatusForbidden, constants.CodeForbidden, constants.MsgForbidden)
		default:
			logger.Error("unhandled error", "error", err)
			util.Fail(c, http.StatusInternalServerError, constants.CodeInternal, constants.MsgInternalError)
		}
	}
}
