package dm_http

import (
	"dagproc/internal/di"
	"dagproc/internal/di_modplug"
	"errors"
	"fmt"
	"regexp"
)

func httpConfigKWDF_module(v interface{}, hcp HttpConfigPtr) error {
	switch v := v.(type) {
	case string:
		if v != "http" {
			return fmt.Errorf("module should be 'http' instead of %s", v)
		}
		return nil
	default:
		return fmt.Errorf("'module' shoudl be string")
	}
}
func httpConfigKWDF_hooks(v interface{}, hcp HttpConfigPtr) error {
	switch v := v.(type) {
	case []interface{}:
		return di_modplug.ValidateHooks(v, "http")
	default:
		return errors.New("'hooks' must be a list of hooks")
	}
}
func httpConfigKWDF_uri(v interface{}, hcp HttpConfigPtr) error {
	switch v := v.(type) {
	case string:
		var m bool
		if m, _ = regexp.Match("https?://.*(:[0-9]+)?/?.*", []byte(v)); m { // host and port
			hcp.uri = v
			return nil
		}
		return fmt.Errorf("%s does not look like http uri (does not match 'https?://.*(:[0-9]+)?/?.*')", v)
	default:
		return errors.New("'uri' should be string 'https?://.*(:[0-9]+)?/?.*")
	}
}
func httpConfigKWDF_listen(v interface{}, hcp HttpConfigPtr) error {
	switch v := v.(type) {
	case string:
		var m bool
		if m, _ = regexp.Match(".*:[0-9]{1,5}$", []byte(v)); m {
			hcp.listen = v
			return nil
		}
		return fmt.Errorf("invalid host:port specification for 'listen': %s ", v)
	default:
		return errors.New("'listen' should be string 'hostname:port'")
	}
}

func httpConfigKWDF_topic(v interface{}, hcp HttpConfigPtr) error {
	switch v := v.(type) {
	case string:
		hcp.topic = v
		return nil
	default:
		return errors.New("'topic' must be string")
	}
}

func httpLoadConfig(config di.CFConfig, isGlobal bool) (di.PluginConfig, error) {
	var err error
	var ret HttpConfig
	var k string
	var v interface{}
	var kwdfm map[string]HttpConfigKWD = map[string]HttpConfigKWD{
		"module": {dispFunc: httpConfigKWDF_module, dispFlags: di.CKW_GLOBAL},
		"hooks":  {dispFunc: httpConfigKWDF_hooks, dispFlags: di.CKW_GLOBAL},
		"uri":    {dispFunc: httpConfigKWDF_uri, dispFlags: di.CKW_GLOBAL},
		"listen": {dispFunc: httpConfigKWDF_listen, dispFlags: di.CKW_GLOBAL},
		"topic":  {dispFunc: httpConfigKWDF_topic, dispFlags: di.CKW_CHANNEL},
	}
	for k, v = range config {
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
		if (ret.uri == "") && (ret.listen == "") {
			err = errors.New("missing either 'uri' or 'listen' keyword")
			mLog.Error().Err(err).Send()
			return ret, err
		}
	}
	mLog.Debug().Msgf("Loaded config: %+v", ret)
	return ret, nil
}
