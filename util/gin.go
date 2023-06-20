package util

import (
	"github.com/bytedance/go-tagexpr/v2/binding"
	"github.com/gin-gonic/gin"
)

// BindAndValidate binds data from *gin.Context to obj and validates them if needed.
func BindAndValidate(c *gin.Context, obj interface{}) error {
	return binding.BindAndValidate(obj, c.Request, c.Params)
}
