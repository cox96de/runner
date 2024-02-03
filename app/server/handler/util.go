package handler

import (
	"github.com/bytedance/go-tagexpr/v2/binding"
	"github.com/gin-gonic/gin"
)

// Bind binds the request to the object.
func Bind(c *gin.Context, obj interface{}) error {
	return binding.Bind(obj, c.Request, c.Params)
}
