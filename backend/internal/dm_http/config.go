package dm_http

import (
	"dagproc/internal/di"
	"errors"
	"fmt"
	"regexp"
)

func httpConfigKWDF_uri(v interface{}, hcp HttpConfigPtr) error {
	var m bool
	if m, _ = regexp.Match("https?://.*(:[0-9]+)?/?.*", []byte(v.(string))); m { // host and port
		hcp.uri = v.(string)
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

func httpConfigKWDF_topic(v interface{}, hcp HttpConfigPtr) error {
	hcp.topic = v.(string)
	return nil
}

func httpConfigKWDF_template(v interface{}, hcp HttpConfigPtr) error {
	hcp.template = v.(string)
	return nil
}

func httpLoadConfig(config interface{}, isGlobal bool, path string) (di.PluginConfig, error) {
	var err error
	var ret HttpConfig
	var k string
	var v interface{}
	var kwdfm map[string]HttpConfigKWD = map[string]HttpConfigKWD{
		"uri":    {dispFunc: httpConfigKWDF_uri, dispFlags: di.CKW_GLOBAL},
		"listen": {dispFunc: httpConfigKWDF_listen, dispFlags: di.CKW_GLOBAL},
		"topic":  	{dispFunc: httpConfigKWDF_topic, dispFlags: di.CKW_CHANNEL},
		"template":	{dispFunc: httpConfigKWDF_template, dispFlags: di.CKW_CHANNEL|di.CKW_CHANNEL},
	}
	err = di.ValidateConfig(` { "uri": "string", "listen": "string", "topic": "string", "template": "string"}`, config, path)
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
		if (ret.uri == "") && (ret.listen == "") {
			err = errors.New("missing either 'uri' or 'listen' keyword")
			mLog.Error().Err(err).Send()
			return ret, err
		}
	}
	mLog.Debug().Msgf("Loaded config: %+v", ret)
	return ret, nil
}
