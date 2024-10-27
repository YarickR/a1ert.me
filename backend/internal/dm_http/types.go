package dm_http

import "dagproc/internal/di"

type HttpConfigKWDF func(v interface{}, hcp HttpConfigPtr) error // kw == keyword, df == dispatcher func
type HttpConfigKWD struct {
	dispFunc  HttpConfigKWDF
	dispFlags uint
}
type HttpConfigPtr *HttpConfig
type HttpConfig struct {
	server   di.CoVa[string]
	path     di.CoVa[string]
	listen   di.CoVa[string]
	method   di.CoVa[string]
	hdrtmpl  di.CoVa[string]
	bodytmpl di.CoVa[string]
}
