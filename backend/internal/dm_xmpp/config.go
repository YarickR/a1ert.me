package dm_xmpp
import (
	"fmt"
	"errors"
    "dagproc/internal/di"
    "dagproc/internal/di_modplug"
)

func xmppLoadConfig(config di.CFConfig) (di.PluginConfig, error) {
    var err error
    var ret XmppConfig
    var k string
    var v interface{}
    var kwdfm map[string]XmppConfigKWDF = map[string]XmppConfigKWDF {
        "module":		xmppConfigKWDF_module,
        "hooks":		xmppConfigKWDF_hooks,
        "server":		xmppConfigKWDF_server,
        "login":		xmppConfigKWDF_login,
        "password":		xmppConfigKWDF_password,
        "groupsURI":	xmppConfigKWDF_groupsURI,

    }
    for k, v = range config {
        err = kwdfm[k](v, &ret)
        if (err != nil) {
            mLog.Error().Str("keyword", k).Err(err).Send()
            return ret, err
        }
    }
    if ((ret.server == "") || (ret.login == "") || (ret.password == "")) {
        err = errors.New("Missing 'server', 'login' or 'password' keyword")
        mLog.Error().Err(err).Send()
        return ret, err
    }
    mLog.Debug().Msgf("Loaded config: %v", ret)
    return ret, nil
}

func xmppConfigKWDF_module (v interface{}, xcp XmppConfigPtr) error {
	switch t := v.(type) {
		case string:
			if (v.(string) != "xmpp") {
				return errors.New(fmt.Sprintf("module should be 'xmpp' instead of %s", v.(string)))
			}
			return nil
		default:
			return errors.New(fmt.Sprintf("Wrong type for 'module' keyword: %T, should be 'string'", t))
	}
}
func xmppConfigKWDF_hooks (v interface{}, xcp XmppConfigPtr) error {
	switch t := v.(type) {
		case []interface{}:
			return di_modplug.ValidateHooks(v.([]interface{}), "xmpp")
		default:
			return errors.New(fmt.Sprintf("Wrong type for 'hooks' keyword: %T", t))
	}
}
func xmppConfigKWDF_server (v interface{}, xcp XmppConfigPtr) error {
	switch t := v.(type) {
		case string:
			xcp.server = v.(string)
			return nil
		default:
			return errors.New(fmt.Sprintf("Wrong type for 'server' keyword: %T, should be string ", t))
	}
}
func xmppConfigKWDF_login (v interface{}, xcp XmppConfigPtr) error {
	switch t := v.(type) {
		case string:
			xcp.login = v.(string)
			return nil
		default:
			return errors.New(fmt.Sprintf("Wrong type for 'login' keyword: %T, should be string ", t))
	}
}
func xmppConfigKWDF_password (v interface{}, xcp XmppConfigPtr) error {
	switch t := v.(type) {
		case string:
			xcp.password = v.(string)
			return nil
		default:
			return errors.New(fmt.Sprintf("Wrong type for 'password' keyword: %T, should be string ", t))
	}
}

func xmppConfigKWDF_groupsURI (v interface{}, xcp XmppConfigPtr) error {
	switch t := v.(type) {
		case string:
			xcp.groupsURI = v.(string)
			return nil
		default:
			return errors.New(fmt.Sprintf("Wrong type for 'groupsURI' keyword: %T, should be string ", t))
	}
}
