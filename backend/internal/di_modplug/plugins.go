package di_modplug

import (
	"errors"
	"fmt"
	"dagproc/internal/di"
// 	"github.com/rs/zerolog/log"
)

func LoadPluginsConfig(cfg di.CFConfig, mods map[string]di.Module) (map[string]di.PluginPtr, error) {
	var err error // pcle == plugin config load error
	var ok bool // generic ok
	var pn, mn string //plugin name, module name
	var pcd interface{} // pcd == plugin config description 
	var ret map[string]di.PluginPtr
	
	if (cfg == nil) {
		err = errors.New("Empty 'plugins' config section")
		return nil, err
	}
	ret = make(map[string]di.PluginPtr)

	ret["core"] = &di.Plugin {
		Name: 	"core",
		Type: 	di.PT_PROC,
		Module: mods["core"],
		Config: nil,
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
		var rmi di.PluginPtr  // ret map item
		rmi = &di.Plugin {
			Module: mod,
		} 
		rmi.Config, err = mod.Hooks.LoadConfigHook(pc)
		if (err != nil) {
			return nil, err
		}
		ret[pn] = rmi

	}
	return ret, nil
}
