package gohttpparam

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/gokits/gotools"
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

//copy from go1.7.4 src code
func tagLookup(tag reflect.StructTag, key string) (value string, ok bool) {
	// When modifying this code, also update the validateStructTag code
	// in golang.org/x/tools/cmd/vet/structtag.go.

	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		if key == name {
			value, err := strconv.Unquote(qvalue)
			if err != nil {
				break
			}
			return value, true
		}
	}
	return "", false
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

		tagstr, ok := tagLookup(t.Tag, "param")
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
