package httpparam

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/rs/xmux"
)

type HttpparamError error

var (
	TagInvalid HttpparamError = errors.New("TagInvalid")
)

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

func decodeTag(tag string, t *tagInfo) (bool, error) {
	tags := strings.Split(tag, " ")
	t.Required = false
	for _, tf := range tags {
		fs := strings.Split(t, "=")
		if fs[0] == "query" {
			if len(fs) != 2 {
				return false, TagInvalid
			}
			t.Type = Query
			t.Name = fs[1]
			return true, nil
		} else if fs[0] == "path" {
			if len(fs) != 2 {
				return false, TagInvalid
			}
			t.Type = Path
			t.Name = fs[1]
			return true, nil
		} else if fs[0] == "required" {
			if len(fs) != 1 {
				return false, TagInvalid
			}
			t.Required = true
			return true, nil
		}
	}
	return false, nil
}

func NewParamHttpC(f func(ctx context.Context, w http.ResponseWriter, r *http.Request), parampool *sync.Pool) xhandler.HandlerFuncC {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		params := parampool.Get()
		values := reflect.ValueOf(params)
		if values.Kind() != reflect.Ptr {
			panic("param must be ptr")
		}
		values = values.Elem()
		fields := gotools.DeepFields(values.Type())
		var taginfo tagInfo
		for _, v := range fields {
			var pv interface{}
			tagstr, ok := v.Tag.Lookup("param")
			if !ok {
				continue
			}
			ok = decodeTag(tagstr, &taginfo)
			if !ok {
				continue
			}
			var pvs string
			if taginfo.Type == Query {
				qvs, ok := r.URL.Query()[taginfo.Name]
				if !ok {
					if taginfo.Required {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte{"QueryString param "})
						w.Write([]byte{taginfo.Name})
						w.Write([]byte{" needed"})
						return
					}
					continue
				}
				pvs = qvs[0]
			} else {
				pvs = xmux.Param(ctx, taginfo.Name)
			}
			switch v.Type.Kind() {
			case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
				pv := strings
			}
		}
	}
}
