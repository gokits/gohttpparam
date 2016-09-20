package gohttpparam

import (
	"testing"
)

func TestDecodeTagQuery(t *testing.T) {
	var ti tagInfo
	err := decodeTag("a", "query=a", &ti)
	if err != nil {
		t.Errorf("decode a failed")
	}
	if ti.Name != "a" || ti.Required || ti.Type != Query {
		t.Errorf("decode a incorrect")
	}
}

func TestDecodeTagPath(t *testing.T) {
	var ti tagInfo
	err := decodeTag("b", "path=b", &ti)
	if err != nil {
		t.Errorf("decode b failed")
	}
	if ti.Name != "b" || ti.Required || ti.Type != Path {
		t.Errorf("decode b incorrect, taginfo=%v", ti)
	}
}

func TestDecodeTagRequired(t *testing.T) {
	var ti tagInfo
	err := decodeTag("b", "path=b,required", &ti)
	if err != nil {
		t.Errorf("decode b failed")
	}
	if ti.Name != "b" || !ti.Required || ti.Type != Path {
		t.Errorf("decode b incorrect, taginfo=%v", ti)
	}
}

func TestDecodeTagErrNotFound(t *testing.T) {
	var ti tagInfo
	err := decodeTag("b", "required", &ti)
	if err == nil {
		t.Errorf("error expected")
	}
	_, ok := err.(*ErrTagFieldNotFound)
	if !ok {
		t.Errorf("ErrTagFieldNotFound expected")
	}
}

var paths = map[string]string{
	"b": "1",
	"d": "1.242334353",
}

var querys = map[string]string{
	"a": "hahaha",
	"c": "33",
}

type param struct {
	A *string  `param:"query=a"`
	B *bool    `param:"path=b,required"`
	C *int8    `param:"query=c"`
	D *float32 `param:"path=d"`
	E int
	F *uint8 `param:"path=f"`
	G *int   `param:"path=g,default=3"`
}

type notSupported struct {
	A map[int]int `param:"query=a"`
}

func TestDecodeParamsNormal(t *testing.T) {
	var p param
	err := DecodeParams(
		&p,
		func(key string) (v string, ok bool) {
			v, ok = paths[key]
			return
		},
		func(key string) (v string, ok bool) {
			v, ok = querys[key]
			return
		},
	)
	if err != nil {
		t.Errorf("DecodeParams failed. err=%v", err)
	}
	if p.A == nil || *p.A != "hahaha" || p.B == nil || *p.B != true || p.C == nil || *p.C != 33 || p.D == nil || *p.D != 1.242334353 || p.E != 0 || p.F != nil {
		t.Errorf("DecodeParams incorrect")
	}
	if *p.G != 3 {
		t.Errorf("default decode failed")
	}
}

func TestDecodeParamsErrNotSupported(t *testing.T) {
	var p notSupported
	err := DecodeParams(
		&p,
		func(key string) (v string, ok bool) {
			v, ok = paths[key]
			return
		},
		func(key string) (v string, ok bool) {
			v, ok = querys[key]
			return
		},
	)
	if err == nil {
		t.Errorf("DecodeParams error expected. ")
	}
	_, ok := err.(*ErrTypeNotSupported)
	if !ok {
		t.Errorf("DecodeParams ErrTypeNotSupported expected")
	}
}
