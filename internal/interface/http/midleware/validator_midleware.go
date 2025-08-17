package midleware

import (
	"fmt"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"reflect"
)

var Validator = validator.New()

type CustomValidatable interface {
	Validate() error
}

func init() {
	Validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("json")
		if name == "-" {
			return ""
		}
		return name
	})
}

func EnsureJsonValidRequest[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		body := new(T)

		if err := c.ShouldBindJSON(body); err != nil {
			logrus.Warning("Failed to bind JSON: ", err)
			response.BadRequest(c, "invalid request", err)
			c.Abort()
			return
		}

		if err := Validator.Struct(body); err != nil {
			var errStr string
			for i, e := range err.(validator.ValidationErrors) {
				if i > 0 {
					errStr += ", "
				}
				errStr += fmt.Sprintf("%s %s", e.Field(), e.Tag())
			}
			logrus.Warning("Validation errors: ", errStr)
			response.BadRequest(c, "invalid request", fmt.Errorf(errStr))
			c.Abort()
			return
		}

		if custom, ok := any(body).(CustomValidatable); ok {
			if err := custom.Validate(); err != nil {
				logrus.Warning("Custom validation failed: ", err)
				response.BadRequest(c, "invalid request", err)
				c.Abort()
				return
			}
		}

		c.Set("body", body)

		c.Next()
	}
}

func EnsureMultipartValidRequest[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		body := new(T)

		if err := c.ShouldBind(body); err != nil {
			response.BadRequest(c, "invalid request", err)
			c.Abort()
			return
		}

		if err := Validator.Struct(body); err != nil {
			var errStr string
			for i, e := range err.(validator.ValidationErrors) {
				if i > 0 {
					errStr += ", "
				}
				errStr += fmt.Sprintf("%s %s", e.Field(), e.Tag())
			}
			response.BadRequest(c, "invalid request", fmt.Errorf(errStr))
			c.Abort()
			return
		}

		if custom, ok := any(body).(CustomValidatable); ok {
			if err := custom.Validate(); err != nil {
				response.BadRequest(c, "invalid request", err)
				c.Abort()
				return
			}
		}

		c.Set("body", body)
		c.Next()
	}
}
