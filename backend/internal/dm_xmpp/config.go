package dm_xmpp

import (
	"dagproc/internal/di"
	"dagproc/internal/di_modplug"
	"errors"
	"fmt"
)

func xmppLoadConfig(config interface{}, isGlobal bool, path string) (di.PluginConfig, error) {
	var err error
	var f XmppConfigKWD
	var ok bool
	var ret XmppConfig
	var k string
	var v interface{}
	var kwdfm map[string]XmppConfigKWD = map[string]XmppConfigKWD{
		"module":    {dispFunc: xmppConfigKWDF_module, dispFlags: di.CKW_GLOBAL},
		"hooks":     {dispFunc: xmppConfigKWDF_hooks, dispFlags: di.CKW_GLOBAL},
		"server":    {dispFunc: xmppConfigKWDF_server, dispFlags: di.CKW_GLOBAL},
		"login":     {dispFunc: xmppConfigKWDF_login, dispFlags: di.CKW_GLOBAL},
		"password":  {dispFunc: xmppConfigKWDF_password, dispFlags: di.CKW_GLOBAL},
		"groupsURI": {dispFunc: xmppConfigKWDF_groupsURI, dispFlags: di.CKW_GLOBAL},
		"template":  {dispFunc: xmppConfigKWDF_template, dispFlags: di.CKW_GLOBAL | di.CKW_CHANNEL},
		"group":     {dispFunc: xmppConfigKWDF_group, dispFlags: di.CKW_CHANNEL},
	}
	err = di.ValidateConfig(`
		{ 	"module!": "string", 
			"hooks!": [], 
			"server": 		"string", 
			"login": 		"string", 
			"password": 	"string", 
			"groupsURI":	"string",
			"template": 	"string", 
			"group": 		"string" 
		}
	`, config, path)
	if err != nil {
		return ret, err
	}
	for k, v = range config.(di.MSI) {
		f, ok = kwdfm[k]
		if !ok {
			err = fmt.Errorf("Unknown keyword '%s'", k)
			mLog.Error().Str("keyword", k).Err(err).Send()
			return ret, err
		}
		if isGlobal && ((f.dispFlags & di.CKW_GLOBAL) == 0) {
			err = fmt.Errorf("'%s' cannot be used in global config", k)
			mLog.Error().Str("keyword", k).Err(err).Send()
			return ret, err
		}
		if !isGlobal && ((f.dispFlags & di.CKW_CHANNEL) == 0) {
			err = fmt.Errorf("'%s' cannot be used in per-channel config", k)
			mLog.Error().Str("keyword", k).Err(err).Send()
			return ret, err
		}

		err = f.dispFunc(v, &ret)
		if err != nil {
			mLog.Error().Str("keyword", k).Err(err).Send()
			return ret, err
		}
	}
	if isGlobal {
		if (ret.server == "") || (ret.login == "") || (ret.password == "") {
			err = errors.New("Missing 'server', 'login' or 'password' keyword")
			mLog.Error().Err(err).Send()
			return ret, err
		}
	}
	mLog.Debug().Msgf("Loaded config: %+v", ret)
	return ret, nil
}

func xmppConfigKWDF_module(v interface{}, xcp XmppConfigPtr) error {
	switch v.(type) {
	case string:
		if v.(string) != "xmpp" {
			return errors.New(fmt.Sprintf("module should be 'xmpp' instead of %s", v.(string)))
		}
		return nil
	default:
		return errors.New(fmt.Sprintf("'module' should be 'string'"))
	}
}
func xmppConfigKWDF_hooks(v interface{}, xcp XmppConfigPtr) error {
	switch v.(type) {
	case []interface{}:
		return di_modplug.ValidateHooks(v.([]interface{}), "xmpp")
	default:
		return errors.New(fmt.Sprintf("'hooks' should be a list of active hooks for this plugin"))
	}
}
func xmppConfigKWDF_server(v interface{}, xcp XmppConfigPtr) error {
	switch v.(type) {
	case string:
		xcp.server = v.(string)
		return nil
	default:
		return errors.New(fmt.Sprintf("'server' should be string"))
	}
}
func xmppConfigKWDF_login(v interface{}, xcp XmppConfigPtr) error {
	switch v.(type) {
	case string:
		xcp.login = v.(string)
		return nil
	default:
		return errors.New(fmt.Sprintf("'login' should be string "))
	}
}
func xmppConfigKWDF_password(v interface{}, xcp XmppConfigPtr) error {
	switch v.(type) {
	case string:
		xcp.password = v.(string)
		return nil
	default:
		return errors.New(fmt.Sprintf("'password' should be string "))
	}
}

func xmppConfigKWDF_groupsURI(v interface{}, xcp XmppConfigPtr) error {
	switch v.(type) {
	case string:
		xcp.groupsURI = v.(string)
		return nil
	default:
		return errors.New(fmt.Sprintf("'groupsURI' should be string "))
	}
}

func xmppConfigKWDF_template(v interface{}, xcp XmppConfigPtr) error {
	var ok bool
	switch v := v.(type) {
	case string:
		xcp.template, ok = di.GCfg.Templates[v]
		if !ok {
			xcp.template = nil
			return fmt.Errorf("undefined template '%s'", v)
		}
		return nil
	default:
		return errors.New("'template' should be string ")
	}
}
func xmppConfigKWDF_group(v interface{}, xcp XmppConfigPtr) error {
	switch v := v.(type) {
	case string:
		xcp.group = v
		return nil
	default:
		return errors.New("'groups' should be string ")
	}

}
