package handler

import (
	"net/http"

	"github.com/cox96de/runner/lib"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

var jj = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {
	jj.RegisterExtension(&lib.ProtobufTypeExtension{})
}

type render struct {
	data interface{}
}

func (r *render) Render(writer http.ResponseWriter) error {
	bs, err := jj.Marshal(r.data)
	if err != nil {
		return err
	}
	_, err = writer.Write(bs)
	return err
}

func (r *render) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, []string{"application/json; charset=utf-8"})
}

func JSON(c *gin.Context, code int, data interface{}) {
	c.Render(code, &render{data: data})
}

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}
