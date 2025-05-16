package midleware

import (
	"fmt"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"reflect"
)

var Validator = validator.New()

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
			response.BadRequest(c, "invalid request", err)
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
			return
		}

		c.Set("body", body)

		c.Next()
	}
}
