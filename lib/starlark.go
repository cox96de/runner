package lib

import (
	"reflect"

	"github.com/pkg/errors"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const tagKey = "json"

// ConvertToStarlarkValue converts a Go value to a Starlark value.
func ConvertToStarlarkValue(a interface{}) (starlark.Value, error) {
	value := reflect.ValueOf(a)
	switch value.Kind() {
	case reflect.String:
		return starlark.String(value.String()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
		reflect.Uint16, reflect.Uint32:
		return starlark.MakeInt(int(value.Int())), nil
	case reflect.Bool:
		return starlark.Bool(value.Bool()), nil
	case reflect.Struct:
		return convertToStarlarkStruct(a)
	case reflect.Float64, reflect.Float32:
		return starlark.Float(value.Float()), nil
	case reflect.Array, reflect.Slice:
		eles := make([]starlark.Value, 0, value.Len())
		for i := 0; i < value.Len(); i++ {
			rv, err := ConvertToStarlarkValue(value.Index(i).Interface())
			if err != nil {
				return nil, errors.WithMessagef(err, "failed to convert value for index %d", i)
			}
			eles = append(eles, rv)
		}
		return starlark.NewList(eles), nil
	case reflect.Map:
		rm := starlark.NewDict(value.Len())
		for _, k := range value.MapKeys() {
			v := value.MapIndex(k)
			rv, err := ConvertToStarlarkValue(v.Interface())
			if err != nil {
				return nil, errors.WithMessagef(err, "failed to convert value for key %v", k.Interface())
			}
			kv, err := ConvertToStarlarkValue(k.Interface())
			if err != nil {
				return nil, errors.WithMessagef(err, "failed to convert key %v", k.Interface())
			}
			if kv == nil {
				continue
			}
			err = rm.SetKey(kv, rv)
			if err != nil {
				return nil, errors.WithMessagef(err, "failed to set key %v", k.Interface())
			}
		}
		return rm, nil
	case reflect.Ptr:
		if value.IsValid() && !value.IsNil() {
			return ConvertToStarlarkValue(value.Elem().Interface())
		}
		return nil, nil
	default:
		return nil, errors.Errorf("unsupported value type: %s", value.Type())
	}
}

func convertToStarlarkStruct(a interface{}) (*starlarkstruct.Struct, error) {
	v := reflect.ValueOf(a)
	if v.Kind() != reflect.Struct {
		return nil, errors.Errorf("input should be a struct type")
	}
	t := v.Type()
	dict := make(starlark.StringDict)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		// Convert the Go value to a Starlark value
		var (
			starlarkValue starlark.Value
			err           error
		)
		fieldName := field.Name
		if tag, ok := field.Tag.Lookup(tagKey); ok {
			fieldName = tag
		}
		starlarkValue, err = ConvertToStarlarkValue(value.Interface())
		if err != nil {
			return nil, errors.WithMessagef(err, "failed to convert field %s(%s)", fieldName, field.Name)
		}
		if starlarkValue == nil {
			continue
		}
		dict[fieldName] = starlarkValue
	}
	return starlarkstruct.FromStringDict(starlarkstruct.Default, dict), nil
}
