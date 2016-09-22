package gohttpparam

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/YueHonghui/gotools"
)

type ErrTagInvalid struct {
	Field string
}

func (e *ErrTagInvalid) Error() string {
	return "tag of field " + e.Field + " invalid"
}

type ErrValueInvalid struct {
	Param string
}

func (e *ErrValueInvalid) Error() string {
	return "param " + e.Param + " invalid"
}

type ErrParamRequired struct {
	Param string
}

func (e *ErrParamRequired) Error() string {
	return "param " + e.Param + " required"
}

type ErrTypeNotSupported struct {
	Field string
}

func (e *ErrTypeNotSupported) Error() string {
	return "type of field " + e.Field + " not supported"
}

type ErrTagFieldNotFound struct {
	Field string
}

func (e *ErrTagFieldNotFound) Error() string {
	return "field " + e.Field + " of tag not found"
}

type ParamType int

const (
	Query ParamType = iota
	Path
)

type tagInfo struct {
	Type     ParamType
	Name     string
	Default  *string
	Required bool
}

func decodeTag(field, tag string, t *tagInfo) error {
	errtag := &ErrTagInvalid{
		Field: field,
	}
	tags := strings.Split(tag, ",")
	t.Required = false
	for _, tf := range tags {
		fs := strings.Split(tf, "=")
		if fs[0] == "query" {
			if len(fs) != 2 {
				return errtag
			}
			t.Type = Query
			t.Name = fs[1]
		} else if fs[0] == "path" {
			if len(fs) != 2 {
				return errtag
			}
			t.Type = Path
			t.Name = fs[1]
		} else if fs[0] == "required" {
			if len(fs) != 1 {
				return errtag
			}
			t.Required = true
		} else if fs[0] == "default" {
			if len(fs) != 2 {
				return errtag
			}
			t.Default = &fs[1]
		}
	}
	if t.Name == "" {
		return &ErrTagFieldNotFound{"path/query"}
	}
	return nil
}

func DecodeParams(params interface{}, pathget func(key string) (string, bool), queryget func(key string) (string, bool)) error {
	values := reflect.ValueOf(params)
	if values.Kind() != reflect.Ptr {
		panic("param must be ptr")
	}
	values = values.Elem()
	fields := gotools.DeepFields(values.Type())
	for _, t := range fields {
		if t.Anonymous {
			continue
		}
		fv := values.FieldByName(t.Name)
		if !fv.IsValid() || !fv.CanSet() {
			continue
		}
		tagstr, ok := t.Tag.Lookup("param")
		if !ok {
			continue
		}
		var taginfo tagInfo
		err := decodeTag(t.Name, tagstr, &taginfo)
		if err != nil {
			return err
		}
		var pvs string
		if taginfo.Type == Query {
			pvs, ok = queryget(taginfo.Name)
		} else {
			pvs, ok = pathget(taginfo.Name)
		}
		if !ok {
			if !taginfo.Required && taginfo.Default == nil {
				continue
			}
			if taginfo.Required {
				return &ErrParamRequired{t.Name}
			}
			if taginfo.Default != nil {
				pvs = *taginfo.Default
			}
		}
		vetype := t.Type
		if vetype.Kind() == reflect.Ptr {
			vetype = vetype.Elem()
		}
		switch vetype.Kind() {
		case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
			pv, err := strconv.ParseInt(pvs, 10, vetype.Bits())
			if err != nil {
				if !ok {
					return &ErrTagInvalid{t.Name}
				}
				return &ErrValueInvalid{t.Name}
			}
			if t.Type.Kind() == reflect.Ptr {
				if fv.IsNil() {
					intptr := reflect.New(fv.Type().Elem())
					intptr.Elem().SetInt(pv)
					fv.Set(intptr)
				} else {
					fv.Elem().SetInt(pv)
				}
			} else {
				fv.SetInt(pv)
			}
		case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
			pv, err := strconv.ParseUint(pvs, 10, vetype.Bits())
			if err != nil {
				if !ok {
					return &ErrTagInvalid{t.Name}
				}
				return &ErrValueInvalid{t.Name}
			}
			if t.Type.Kind() == reflect.Ptr {
				if fv.IsNil() {
					integar := reflect.New(fv.Type().Elem())
					integar.Elem().SetUint(pv)
					fv.Set(integar)
				} else {
					fv.Elem().SetUint(pv)
				}
			} else {
				fv.SetUint(pv)
			}
		case reflect.String:
			if t.Type.Kind() == reflect.Ptr {
				if fv.IsNil() {
					strptr := &pvs
					fv.Set(reflect.ValueOf(strptr))
				} else {
					fv.Elem().SetString(pvs)
				}
			} else {
				fv.SetString(pvs)
			}
		case reflect.Bool:
			b, err := strconv.ParseBool(pvs)
			if err != nil {
				if !ok {
					return &ErrTagInvalid{t.Name}
				}
				return &ErrValueInvalid{t.Name}
			}
			if t.Type.Kind() == reflect.Ptr {
				if fv.IsNil() {
					bptr := &b
					fv.Set(reflect.ValueOf(bptr))
				} else {
					fv.Elem().SetBool(b)
				}
			} else {
				fv.SetBool(b)
			}
		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(pvs, vetype.Bits())
			if err != nil {
				if !ok {
					return &ErrTagInvalid{t.Name}
				}
				return &ErrValueInvalid{t.Name}
			}
			if t.Type.Kind() == reflect.Ptr {
				if fv.IsNil() {
					floatptr := reflect.New(vetype)
					floatptr.Elem().SetFloat(f)
					fv.Set(floatptr)
				} else {
					fv.Elem().SetFloat(f)
				}
			} else {
				fv.SetFloat(f)
			}
		default:
			return &ErrTypeNotSupported{t.Name}
		}
	}
	return nil
}
