package form

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

func Unmarshal(r *http.Request, i interface{}) error {
	rv := reflect.ValueOf(i)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &InvalidUnmarshalError{
			Type: reflect.TypeOf(i),
		}

	}

	s := rv.Elem()
	if s.Kind() != reflect.Struct {
		return &InvalidUnmarshalError{
			Type: reflect.TypeOf(i),
		}
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

func Marshal(r *http.Request, i interface{}) error {
	rv := reflect.ValueOf(i)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &InvalidMarshalError{
			Type: reflect.TypeOf(i),
		}
	}

	s := rv.Elem()
	if s.Kind() != reflect.Struct {
		return &InvalidMarshalError{
			Type: reflect.TypeOf(i),
		}
	}

	form := make(url.Values)
	for i := 0; i < s.NumField(); i++ {
		f := s.Type().Field(i)
		tag := f.Tag.Get("form")
		if tag == "" {
			continue
		}
		err := marshalFormValues(tag, s.Field(i), form)
		if err != nil {
			err.Struct = s.Type().Name()
			err.Field = f.Name
			return err
		}
	}

	r.URL.RawQuery = form.Encode()
	return nil
}

type InvalidUnmarshalError struct {
	Type reflect.Type
}

type InvalidMarshalError struct {
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

func (e *InvalidMarshalError) Error() string {
	if e.Type == nil {
		return "form: Marshal(nil)"
	}
	if e.Type.Kind() != reflect.Pointer {
		return "form: Marshal(non-pointer " + e.Type.String() + ")"
	}
	return "form: Marshal(nil " + e.Type.String() + ")"
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

type MarshalTypeError struct {
	Type   reflect.Type
	Value  interface{}
	Struct string
	Field  string
}

func (e *MarshalTypeError) Error() string {
	return fmt.Sprintf("form: cannot marshal %v (%s) of Go struct field %s.%s into form data", e.Value, e.Type, e.Struct, e.Field)
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

func marshalFormValues(tag string, f reflect.Value, form url.Values) *MarshalTypeError {
	if f.Kind() == reflect.Slice || f.Kind() == reflect.Array {
		for i := 0; i < f.Len(); i++ {
			err := marshalFormValue(tag, f.Index(i), form)
			if err != nil {
				err.Type = f.Type()
				err.Field = f.Type().Name()
				return err
			}
		}
		return nil
	}
	return marshalFormValue(tag, f, form)
}

func marshalFormValue(tag string, f reflect.Value, form url.Values) *MarshalTypeError {
	switch f.Kind() {
	case reflect.String:
		form.Add(tag, f.String())
		return nil
	case reflect.Bool:
		form.Add(tag, fmt.Sprintf("%t", f.Bool()))
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		form.Add(tag, fmt.Sprintf("%d", f.Int()))
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		form.Add(tag, fmt.Sprintf("%d", f.Uint()))
		return nil
	case reflect.Float32, reflect.Float64:
		form.Add(tag, fmt.Sprintf("%f", f.Float()))
		return nil
	case reflect.Complex64, reflect.Complex128:
		form.Add(tag, fmt.Sprintf("%e", f.Complex()))
		return nil
	default:
		return &MarshalTypeError{
			Type:  f.Type(),
			Value: f.Interface(),
		}
	}
}
