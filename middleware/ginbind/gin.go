package ginbind

import (
	"github.com/gin-gonic/gin"
	"github.com/gokits/gohttpparam"
	"github.com/go-validator/validator"
)

func BindParam(ctx *gin.Context, param interface{}) (err error) {
	err = gohttpparam.DecodeParams(param,
		ctx.Params.Get,
		func(key string) (v string, ok bool) {
			vs, oki := ctx.Request.URL.Query()[key]
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
			panic(err)
		default:
			return
		}
	}
	return nil
}
