package gohttpparam

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/YueHonghui/gotools"
	"github.com/rs/xhandler"
	//	"github.com/rs/xmux"
	"golang.org/x/net/context"
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

type ParamType int

const (
	Query ParamType = iota
	Path
)

type tagInfo struct {
	Type     ParamType
	Name     string
	Required bool
}

func decodeTag(field, tag string, t *tagInfo) (bool, error) {
	errtag := &ErrTagInvalid{
		Field: field,
	}
	tags := strings.Split(tag, " ")
	t.Required = false
	for _, tf := range tags {
		fs := strings.Split(tf, "=")
		if fs[0] == "query" {
			if len(fs) != 2 {
				return false, errtag
			}
			t.Type = Query
			t.Name = fs[1]
			return true, nil
		} else if fs[0] == "path" {
			if len(fs) != 2 {
				return false, errtag
			}
			t.Type = Path
			t.Name = fs[1]
			return true, nil
		} else if fs[0] == "required" {
			if len(fs) != 1 {
				return false, errtag
			}
			t.Required = true
			return true, nil
		}
	}
	return false, nil
}

func DecodeParams(params interface{}, pathget func(key string) (string, bool), queryget func(key string) (string, bool)) error {
	values := reflect.ValueOf(params)
	if values.Kind() != reflect.Ptr {
		panic("param must be ptr")
	}
	values = values.Elem()
	fields := gotools.DeepFields(values.Type())
	var taginfo tagInfo
	for _, t := range fields {
		if t.Anonymous {
			continue
		}
		fv := values.FieldByName(t.Name)
		if !fv.IsValid() {
			continue
		}
		var pv interface{}
		tagstr, ok := t.Tag.Lookup("param")
		if !ok {
			continue
		}
		ok, err := decodeTag(t.Name, tagstr, &taginfo)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		var pvs string
		if taginfo.Type == Query {
			pvs, ok = queryget(taginfo.Name)
		} else {
			pvs, ok = pathget(taginfo.Name)
		}
		if !ok {
			if taginfo.Required {
				return &ErrParamRequired{t.Name}
			}
			continue
		}
		vetype := t.Type
		if vetype.Kind() == reflect.Ptr {
			vetype = vetype.Elem()
		}
		switch vetype.Kind() {
		case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
			pv, err := strconv.ParseInt(pvs, 10, 64)
			if err != nil {
				return &ErrValueInvalid{t.Name}
			}
			if pv < -(1<<(uint64(vetype.Bits())-1)) || pv > (1<<(uint64(vetype.Bits())-1))-1 {
				return &ErrValueInvalid{t.Name}
			}
			if t.Type.Kind() == reflect.Ptr {

			}
		}
	}
}

func NewParamHttpC(f func(ctx context.Context, w http.ResponseWriter, r *http.Request), parampool *sync.Pool) xhandler.HandlerFuncC {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		params := parampool.Get()

	}
}
