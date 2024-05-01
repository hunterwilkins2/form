package form

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "form: Unmarshal(nil)"
	}
	if e.Type.Kind() != reflect.Pointer {
		return "form: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "form: Unmarshal(nil " + e.Type.String() + ")"
}

type UnmarshalTypeError struct {
	Value  string
	Type   reflect.Type
	Struct string
	Field  string
	Err    error
}

func (e *UnmarshalTypeError) Error() string {
	return fmt.Sprintf("form: cannot unmarshal %s into Go struct field %s.%s of type %s: %s",
		e.Value, e.Struct, e.Field, e.Type, e.Err)
}

func (e *UnmarshalTypeError) Unwrap() error {
	return e.Err
}

func Unmarshal(r *http.Request, i interface{}) error {
	rv := reflect.ValueOf(i)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(i)}
	}

	s := rv.Elem()
	if s.Kind() != reflect.Struct {
		return &InvalidUnmarshalError{reflect.TypeOf(i)}
	}

	err := r.ParseForm()
	if err != nil {
		return err
	}

	for i := 0; i < s.NumField(); i++ {
		f := s.Type().Field(i)
		tag := f.Tag.Get("form")
		values := r.Form[tag]
		err := parseFormValues(s.Field(i), values)
		if err != nil {
			err.Struct = s.Type().Name()
			err.Field = f.Name
			return err
		}
	}

	return nil
}

func parseFormValues(f reflect.Value, values []string) *UnmarshalTypeError {
	if len(values) == 0 || !f.IsValid() || !f.CanSet() {
		return nil
	}

	if f.Kind() == reflect.Slice {
		s := reflect.MakeSlice(f.Type(), len(values), len(values))
		for i, val := range values {
			err := parseFormValue(s.Index(i), val)
			if err != nil {
				err.Value = "[" + strings.Join(values, ", ") + "]"
				err.Type = f.Type()
				return err
			}
		}
		f.Set(s)
		return nil
	}

	if f.Kind() == reflect.Array {
		if f.Len() != len(values) {
			return &UnmarshalTypeError{
				Value: "[" + strings.Join(values, ", ") + "]",
				Type:  f.Type(),
				Err:   fmt.Errorf("cannot use [%d]%s as %s value in struct", len(values), f.Type().Elem(), f.Type()),
			}
		}
		arr := reflect.ArrayOf(len(values), f.Type().Elem())
		s := reflect.New(arr).Elem()
		for i, val := range values {
			err := parseFormValue(s.Index(i), val)
			if err != nil {
				err.Value = "[" + strings.Join(values, ", ") + "]"
				err.Type = f.Type()
				return err
			}
		}
		f.Set(s)
		return nil
	}

	if len(values) != 1 {
		return &UnmarshalTypeError{
			Value: "[" + strings.Join(values, ", ") + "]",
			Type:  f.Type(),
			Err:   fmt.Errorf("cannot unmarshal more than one value for non-slice field"),
		}
	}

	err := parseFormValue(f, values[0])
	if err != nil {
		return err
	}
	return nil
}

func parseFormValue(f reflect.Value, value string) *UnmarshalTypeError {
	switch f.Kind() {
	case reflect.String:
		f.SetString(value)
		return nil
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return &UnmarshalTypeError{
				Value: value,
				Type:  f.Type(),
				Err:   err,
			}
		}
		f.SetBool(v)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return &UnmarshalTypeError{
				Value: value,
				Type:  f.Type(),
				Err:   err,
			}
		}
		if f.OverflowInt(v) {
			return &UnmarshalTypeError{
				Value: value,
				Type:  f.Type(),
				Err:   fmt.Errorf("%s overflows %s value", value, f.Type()),
			}
		}
		f.SetInt(v)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return &UnmarshalTypeError{
				Value: value,
				Type:  f.Type(),
				Err:   err,
			}
		}
		if f.OverflowUint(v) {
			return &UnmarshalTypeError{
				Value: value,
				Type:  f.Type(),
				Err:   fmt.Errorf("%s overflows %s value", value, f.Type()),
			}
		}
		f.SetUint(v)
		return nil
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return &UnmarshalTypeError{
				Value: value,
				Type:  f.Type(),
				Err:   err,
			}
		}
		if f.OverflowFloat(v) {
			return &UnmarshalTypeError{
				Value: value,
				Type:  f.Type(),
				Err:   fmt.Errorf("%s overflows %s value", value, f.Type()),
			}
		}
		f.SetFloat(v)
		return nil
	case reflect.Complex64, reflect.Complex128:
		v, err := strconv.ParseComplex(value, 128)
		if err != nil {
			return &UnmarshalTypeError{
				Value: value,
				Type:  f.Type(),
				Err:   err,
			}
		}
		if f.OverflowComplex(v) {
			return &UnmarshalTypeError{
				Value: value,
				Type:  f.Type(),
				Err:   fmt.Errorf("%s overflows %s value", value, f.Type()),
			}
		}
		f.SetComplex(v)
		return nil
	default:
		return &UnmarshalTypeError{
			Value: value,
			Type:  f.Type(),
			Err:   fmt.Errorf("type %s cannot be unmarshalled from form", f.Type()),
		}
	}
}
