package dm_http

type HttpConfigKWDF func (v interface{}, hcp HttpConfigPtr) error // kw == keyword, df == dispatcher func

type HttpConfigPtr *HttpConfig
type HttpConfig struct {
	uri 	string
	listen 	string
}
