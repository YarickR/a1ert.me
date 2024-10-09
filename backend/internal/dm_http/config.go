package dm_http

import (
	"dagproc/internal/di"
	"errors"
	"fmt"
	"regexp"
)

func httpConfigKWDF_server(v interface{}, hcp HttpConfigPtr) error {
	var m bool
	if m, _ = regexp.Match("https?://.*(:[0-9]+)?/?.*", []byte(v.(string))); m { // host and port
		hcp.server = v.(string)
		return nil
	}
	return fmt.Errorf("%s does not look like http uri (does not match 'https?://.*(:[0-9]+)?/?.*')", v)
}
func httpConfigKWDF_listen(v interface{}, hcp HttpConfigPtr) error {
	var m bool
	if m, _ = regexp.Match(".*:[0-9]{1,5}$", []byte(v.(string))); m {
		hcp.listen = v.(string)
		return nil
	}
	return fmt.Errorf("invalid host:port specification for 'listen': %s ", v)
}

func httpConfigKWDF_method(v interface{}, hcp HttpConfigPtr) error {
	hcp.method = v.(string)
	return nil
}

func httpConfigKWDF_path(v interface{}, hcp HttpConfigPtr) error {
	hcp.path = v.(string)
	return nil
}

func httpConfigKWDF_hdrtmpl(v interface{}, hcp HttpConfigPtr) error {
	hcp.hdrtmpl = v.(string)
	return nil
}

func httpConfigKWDF_bodytmpl(v interface{}, hcp HttpConfigPtr) error {
	hcp.bodytmpl = v.(string)
	return nil
}

func httpLoadConfig(config interface{}, isGlobal bool, path string) (di.PluginConfig, error) {
	var err error
	var ret HttpConfig
	var k string
	var v interface{}
	var kwdfm map[string]HttpConfigKWD = map[string]HttpConfigKWD{
		"server":   {dispFunc: httpConfigKWDF_server, dispFlags: di.CKW_GLOBAL | di.CKW_CHANNEL},
		"path":     {dispFunc: httpConfigKWDF_path, dispFlags: di.CKW_GLOBAL | di.CKW_CHANNEL},
		"method":   {dispFunc: httpConfigKWDF_method, dispFlags: di.CKW_GLOBAL | di.CKW_CHANNEL},
		"listen":   {dispFunc: httpConfigKWDF_listen, dispFlags: di.CKW_GLOBAL},
		"hdrtmpl":  {dispFunc: httpConfigKWDF_hdrtmpl, dispFlags: di.CKW_GLOBAL | di.CKW_CHANNEL},
		"bodytmpl": {dispFunc: httpConfigKWDF_bodytmpl, dispFlags: di.CKW_GLOBAL | di.CKW_CHANNEL},
	}
	err = di.ValidateConfig(` { 
				"server": "string", 
				"path": "string", 
				"listen": "string",  
				"method": "string", 
				"hdrtmpl": "string", 
				"bodytmpl": "string"
			}`, config, path)
	if err != nil {
		return ret, err
	}
	for k, v = range config.(map[string]interface{}) {
		kwd, ok := kwdfm[k]
		if !ok {
			err = fmt.Errorf("unknown keyword '%s'", k)
			mLog.Error().Str("keyword", k).Err(err)
			return ret, err
		}
		err = kwd.dispFunc(v, &ret)
		if err != nil {
			mLog.Error().Str("keyword", k).Err(err).Send()
			return ret, err
		}
	}
	if isGlobal {
		mConfig = ret
	} else {
		// Let's check if we have at least server and method
		t := di.MergeStructs([]string{"server", "path", "method", "listen", "hdrtmpl", "bodytmpl"},
			interface{}(mConfig),
			interface{}(ret))
		ret = t.(HttpConfig)
		if mConfig.listen == "" {
			if (ret.server == "") || (ret.method == "") {
				err = errors.New("Missing 'server', 'listen' or 'method' keyword")
				mLog.Error().Err(err).Send()
				return ret, err
			}
		}
	}
	mLog.Debug().Msgf("Loaded config: %+v", ret)
	return ret, nil
}
