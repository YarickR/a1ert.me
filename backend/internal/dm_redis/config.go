package dm_redis
import (
	"errors"
	"fmt"
	"regexp"
	"dagproc/internal/di"
	"dagproc/internal/di_modplug"
)
func redisConfigKWDF_module (v interface{}, rcp RedisConfigPtr) error {
	switch t := v.(type) {
		case string:
			if (v.(string) != "redis") {
				return errors.New(fmt.Sprintf("module should be 'redis' instead of %s", v.(string)))
			}
			return nil
		default:
			return errors.New(fmt.Sprintf("Wrong type for 'module' keyword: %T, should be 'string'", t))
	}
}
func redisConfigKWDF_hooks (v interface{}, rcp RedisConfigPtr) error {
	switch t := v.(type) {
		case []interface{}:
			return di_modplug.ValidateHooks(v.([]interface{}), "redis")
		default:
			return errors.New(fmt.Sprintf("Wrong type for 'hooks' keyword: %T", t))
	}
}
func redisConfigKWDF_uri (v interface{}, rcp RedisConfigPtr) error {
	switch t := v.(type) {
		case string:
			var m bool
			if m, _ = regexp.Match("redis://.*:([0-9]+)/.*", []byte(v.(string))); m == true  { // host and port
				rcp.uri = v.(string)
				return nil
			}
			if m, _ = regexp.Match("redis:///.+", []byte(v.(string))); m == true { // unix domain socket
				rcp.uri = v.(string)
				return nil
			}
			return errors.New(fmt.Sprintf("%s does not look like redis uri (does not match 'redis://.*:([0-9]+)/.*' or 'redis:///.+')", v.(string)))
		default:
			return errors.New(fmt.Sprintf("Wrong type for 'uri' keyword: %T, should be string 'redis://[user:password@]host:port/[db][?connection parameters]", t))
	}
}
func redisConfigKWDF_list (v interface{}, rcp RedisConfigPtr) error {
	switch t := v.(type) {
		case string:
			var m bool
			if m, _ = regexp.Match("[0-9a-zA-Z_]", []byte(v.(string))); m == true  {
				rcp.list = v.(string)
				return nil
			}
			return errors.New(fmt.Sprintf("%s does not look like list name", v.(string)))
		default:
			return errors.New(fmt.Sprintf("Wrong type for 'list' keyword: %T, should be string with proper redis identifier", t))
	}
}

func redisLoadConfig(config di.CFConfig) (di.PluginConfig, error) {
    var err error
    var ret RedisConfig
    var k string
    var v interface{}
    var f RedisConfigKWDF
    var ok bool
    var kwdfm map[string]RedisConfigKWDF = map[string]RedisConfigKWDF {
    	"module":	redisConfigKWDF_module,
    	"hooks":	redisConfigKWDF_hooks,
    	"uri":		redisConfigKWDF_uri,
    	"list":		redisConfigKWDF_list,
    }
    for k, v = range config {
    	f, ok = kwdfm[k]
    	if (!ok) {
    		err = fmt.Errorf("Unknown keyword '%s'", k);
    		mLog.Error().Err(err).Send()
    		return ret, err
    	}
    	err = f(v, &ret)
    	if (err != nil) {
    		mLog.Error().Str("keyword", k).Err(err).Send()
    		return ret, err
    	}
    }
    if ((ret.uri == "") || (ret.list == "")) {
    	err = errors.New("Missing 'uri' or 'list' keyword")
    	mLog.Error().Err(err).Send()
    	return ret, err
    }
    mLog.Debug().Msgf("Loaded config: %v", ret)
    return ret, nil
}
