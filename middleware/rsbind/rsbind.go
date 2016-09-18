package rsxhandler

import (
	"net/http"

	"github.com/YueHonghui/gohttpparam"
	"github.com/rs/xmux"
	"golang.org/x/net/context"
	"gopkg.in/validator.v2"
)

func BindParam(ctx context.Context, w http.ResponseWriter, r *http.Request, param interface{}) (toContinue bool, err error) {
	err = gohttpparam.DecodeParams(param,
		func(key string) (v string, ok bool) {
			ps := xmux.Params(ctx)
			for _, p := range ps {
				if p.Name == key {
					return p.Value, true
				}
			}
			return "", false
		},
		func(key string) (v string, ok bool) {
			vs, oki := r.URL.Query()[key]
			ok = oki
			if ok {
				v = vs[0]
			}
			return
		},
	)
	if err == nil {
		err = validator.Validate(param)
	}
	if err != nil {
		switch err.(type) {
		case *gohttpparam.ErrTagFieldNotFound, *gohttpparam.ErrTagInvalid, *gohttpparam.ErrTypeNotSupported:
			w.WriteHeader(http.StatusInternalServerError)
			return false, err
		default:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return false, nil
		}
	}
	return true, nil
}
