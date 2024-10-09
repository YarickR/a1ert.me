package dm_http

type HttpConfigKWDF func(v interface{}, hcp HttpConfigPtr) error // kw == keyword, df == dispatcher func
type HttpConfigKWD struct {
	dispFunc  HttpConfigKWDF
	dispFlags uint
}
type HttpConfigPtr *HttpConfig
type HttpConfig struct {
	server   string
	path     string
	listen   string
	method   string
	hdrtmpl  string
	bodytmpl string
}
