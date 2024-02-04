package di_modplug

import (
	"errors"
	"fmt"
	"dagproc/internal/di"
// 	"github.com/rs/zerolog/log"
)


func LoadPluginsConfig(cfg di.CFConfig, mods map[string]di.Module) (map[string]di.Plugin, error) {
	var err error // pcle == plugin config load error
	var ok bool // generic ok
	var pn, mn string //plugin name, module name
	var pcd interface{} // pcd == plugin config description 
	var ret map[string]di.Plugin
	
	if (cfg == nil) {
		err = errors.New("Empty 'plugins' config section")
		return nil, err
	}
	ret = make(map[string]di.Plugin)
	ret["core"] = di.Plugin {
		"core",
		mods["core"],
		nil,
	}
	for pn, pcd = range cfg {
		if _, ok = ret[pn]; ok {
			return nil, fmt.Errorf("Duplicate definition for plugin '%s'", pn)
		}
		switch pcd.(type) {
			case map[string]interface{}:
			default:
				return nil, fmt.Errorf("Invalid config for plugin '%s'", pn)
		}
		var mod di.Module
		var pc di.CFConfig
		pc = pcd.(map[string]interface{})
		if mn, ok = pc["module"].(string); !ok { //mn == module name
			return nil, fmt.Errorf("Missing module name for plugin '%s'", pn)			
		}
		if mod, ok = mods[mn]; !ok {
			return nil, fmt.Errorf("Uknown module '%s' for plugin '%s'", mn, pn)
		}
		var rmi di.Plugin  // ret map item
		rmi.Module =  mod 
		rmi.Config, err = mod.Hooks.LoadConfigHook(pc)
		if (err != nil) {
			return nil, err
		}
		ret[pn] = rmi

	}
	return ret, nil
}
