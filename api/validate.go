package api

import (
	"github.com/go-playground/validator/v10"
)

func ValidateDSL(dsl *PipelineDSL) error {
	vd := validator.New()
	return vd.Struct(dsl)
}
