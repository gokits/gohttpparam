package ginbind

import (
	"net/http"

	"github.com/YueHonghui/gohttpparam"
	"github.com/gin-gonic/gin"
	"gopkg.in/validator.v2"
)

func BindParam(ctx *gin.Context, param interface{}) (toContinue bool, err error) {
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
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return false, err
		default:
			ctx.AbortWithError(http.StatusBadRequest, err)
			return false, nil
		}
	}
	return true, nil
}
