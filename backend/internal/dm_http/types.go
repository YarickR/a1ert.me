package dm_http

type HttpConfigKWDF func(v interface{}, hcp HttpConfigPtr) error // kw == keyword, df == dispatcher func
type HttpConfigKWD struct {
	dispFunc  HttpConfigKWDF
	dispFlags uint
}
type HttpConfigPtr *HttpConfig
type HttpConfig struct {
	uri    string
	listen string
	topic  string
}
