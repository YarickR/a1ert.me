package dm_redis

import (
	"dagproc/internal/di"
	"errors"
	"fmt"
	"regexp"
)

func redisConfigKWDF_uri(v interface{}, rcp RedisConfigPtr) error {
	switch v := v.(type) {
	case string:
		var m bool
		if m, _ = regexp.Match("redis://.*:([0-9]+)/.*", []byte(v)); m == true { // host and port
			rcp.uri = v
			return nil
		}
		if m, _ = regexp.Match("redis:///.+", []byte(v)); m == true { // unix domain socket
			rcp.uri = v
			return nil
		}
		return fmt.Errorf("%s does not look like redis uri (does not match 'redis://.*:([0-9]+)/.*' or 'redis:///.+')", v)
	default:
		return fmt.Errorf("'uri' must be string 'redis://[user:password@]host:port/[db][?connection parameters]")
	}
}
func redisConfigKWDF_list(v interface{}, rcp RedisConfigPtr) error {
	switch v := v.(type) {
		case string:
			var m bool
			if m, _ = regexp.Match("[0-9a-zA-Z_]", []byte(v)); m {
				rcp.list = v
				return nil
			}
			return fmt.Errorf("%s does not look like list name", v)
		default:
			return fmt.Errorf("'list' must be redis identifier")
	}
}

func redisLoadConfig(config interface{}, isGlobal bool, path string) (di.PluginConfig, error) {
	var err error
	var ret RedisConfig
	var k string
	var v interface{}
	var f RedisConfigKWD
	var ok bool
	var kwdfm map[string]RedisConfigKWD = map[string]RedisConfigKWD {
		"uri":    {dispFunc: redisConfigKWDF_uri, dispFlags: di.CKW_GLOBAL},
		"list":   {dispFunc: redisConfigKWDF_list, dispFlags: di.CKW_GLOBAL},
	}
	err = di.ValidateConfig(`{ "uri": "string", "list": "string"}`, config, path)
	if err != nil {
		return ret, err
	}
	for k, v = range config.(map[string]interface{}) {
		f, ok = kwdfm[k]
		if !ok {
			err = fmt.Errorf("unknown keyword '%s'", k)
			mLog.Error().Err(err).Send()
			return ret, err
		}
		err = f.dispFunc(v, &ret)
		if err != nil {
			mLog.Error().Str("keyword", k).Err(err).Send()
			return ret, err
		}
	}
	if (ret.uri == "") || (ret.list == "") {
		err = errors.New("missing 'uri' or 'list' keyword")
		mLog.Error().Err(err).Send()
		return ret, err
	}
	mLog.Debug().Msgf("Loaded config: %+v", ret)
	return ret, nil
}
