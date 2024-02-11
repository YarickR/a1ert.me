package dm_http
import (
	"errors"
	"fmt"
	"regexp"
	"dagproc/internal/di"
	"dagproc/internal/di_modplug"
)
func httpConfigKWDF_module (v interface{}, hcp HttpConfigPtr) error {
	switch t := v.(type) {
		case string:
			if (v.(string) != "http") {
				return errors.New(fmt.Sprintf("module should be 'http' instead of %s", v.(string)))
			}
			return nil
		default:
			return errors.New(fmt.Sprintf("Wrong type for 'module' keyword: %T, should be 'string'", t))
	}
}
func httpConfigKWDF_hooks (v interface{}, hcp HttpConfigPtr) error {
	switch t := v.(type) {
		case []interface{}:
			return di_modplug.ValidateHooks(v.([]interface{}), "http")
		default:
			return errors.New(fmt.Sprintf("Wrong type for 'hooks' keyword: %T", t))
	}
}
func httpConfigKWDF_uri (v interface{}, hcp HttpConfigPtr) error {
	switch t := v.(type) {
		case string:
			var m bool
			if m, _ = regexp.Match("https?://.*(:[0-9]+)?/?.*", []byte(v.(string))); m == true  { // host and port
				hcp.uri = v.(string)
				return nil
			}
			return errors.New(fmt.Sprintf("%s does not look like http uri (does not match 'https?://.*(:[0-9]+)?/?.*')", v.(string)))
		default:
			return errors.New(fmt.Sprintf("Wrong type for 'uri' keyword: %T, should be string 'https?://.*(:[0-9]+)?/?.*", t))
	}
}
func httpConfigKWDF_listen (v interface{}, hcp HttpConfigPtr) error {
	switch t := v.(type) {
		case string:
			var m bool
			if m, _ = regexp.Match(".*:[0-9]{1,5}$", []byte(v.(string))); m == true  {
				hcp.listen = v.(string)
				return nil
			}
			return errors.New(fmt.Sprintf("No way to listen on %s", v.(string)))
		default:
			return errors.New(fmt.Sprintf("Wrong type for 'listen' keyword: %T, should be string 'hostname:port'", t))
	}
}

func httpLoadConfig(config di.CFConfig) (di.PluginConfig, error) {
    var err error
    var ret HttpConfig
    var k string
    var v interface{}
    var kwdfm map[string]HttpConfigKWDF = map[string]HttpConfigKWDF {
    	"module":	httpConfigKWDF_module,
    	"hooks":	httpConfigKWDF_hooks,
    	"uri":		httpConfigKWDF_uri,
    	"listen":	httpConfigKWDF_listen,
    }
    for k, v = range config {
    	err = kwdfm[k](v, &ret)
    	if (err != nil) {
    		mLog.Error().Str("keyword", k).Err(err).Send()
    		return ret, err
    	}
    }
    if ((ret.uri == "") && (ret.listen == "")) {
    	err = errors.New("Missing either 'uri' or 'listen' keyword")
    	mLog.Error().Err(err).Send()
    	return ret, err
    }
    mLog.Debug().Msgf("Loaded config: %v", ret)
    return ret, nil
}
