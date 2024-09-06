package dm_xmpp

import (
	"dagproc/internal/di"
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
		"server":    {dispFunc: xmppConfigKWDF_server, dispFlags: di.CKW_GLOBAL},
		"login":     {dispFunc: xmppConfigKWDF_login, dispFlags: di.CKW_GLOBAL},
		"password":  {dispFunc: xmppConfigKWDF_password, dispFlags: di.CKW_GLOBAL},
		"groupsURI": {dispFunc: xmppConfigKWDF_groupsURI, dispFlags: di.CKW_GLOBAL},
		"template":  {dispFunc: xmppConfigKWDF_template, dispFlags: di.CKW_GLOBAL | di.CKW_CHANNEL},
		"group":     {dispFunc: xmppConfigKWDF_group, dispFlags: di.CKW_CHANNEL},
	}
	if (isGlobal) {
		err = di.ValidateConfig(`
			{ 	
				"server": 		"string", 
				"login": 		"string", 
				"password": 	"string", 
				"groupsURI":	"string",
				"template": 	"string", 
				"group": 		"string" 
			}
		`, config, path) 
	} else { // per channel config may define only different group and template for outgoing messages
		err = di.ValidateConfig(` 
			{ 
				"template": 	"string", 
				"group": 		"string" 
			}`, config, path)
	}
	if err != nil {
		return ret, err
	}
	for k, v = range config.(map[string]interface{}) {
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

func xmppConfigKWDF_server(v interface{}, xcp XmppConfigPtr) error {
	xcp.server = v.(string)
	return nil
}
func xmppConfigKWDF_login(v interface{}, xcp XmppConfigPtr) error {
	xcp.login = v.(string)
	return nil
}
func xmppConfigKWDF_password(v interface{}, xcp XmppConfigPtr) error {
	xcp.password = v.(string)
	return nil
}

func xmppConfigKWDF_groupsURI(v interface{}, xcp XmppConfigPtr) error {
	xcp.groupsURI = v.(string)
	return nil
}

func xmppConfigKWDF_template(v interface{}, xcp XmppConfigPtr) error {
	var ok bool
	xcp.template, ok = di.GCfg.Templates[v.(string)]
	if !ok {
		xcp.template = nil
		return fmt.Errorf("undefined template '%s'", v.(string))
	}
	return nil
}
func xmppConfigKWDF_group(v interface{}, xcp XmppConfigPtr) error {
	xcp.group = v.(string)
	return nil
}
