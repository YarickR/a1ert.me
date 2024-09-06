package di_modplug

import (
	"fmt"
	"dagproc/internal/di"
// 	"github.com/rs/zerolog/log"
)

func LoadPluginsConfig(cfg map[string]interface{}, path string) (map[string]di.PluginPtr, error) {
	var err error // pcle == plugin config load error
	var ok bool // generic ok
	var pn, mn string //plugin name, module name, new path
	var pcd interface{} // pcd == plugin config description 
	var ret map[string]di.PluginPtr
	err = di.ValidateConfig(`
		{
			"*": {
				"module!": "string",
				"type!": "string"
			}
		}`, cfg, path)
	if err != nil	{
		return nil, err
	}
	ret = make(map[string]di.PluginPtr)

	ret["core"] = &di.Plugin {
		Name: 	"core",
		Type: 	"proc",
		Module: di.ModMap["core"],
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
		var (
			mod di.Module
			pc map[string]interface{}
		)
		pc = pcd.(map[string]interface{})
		mn =  pc["module"].(string)
		if mod, ok = di.ModMap[mn]; !ok {
			return nil, fmt.Errorf("Uknown module '%s' for plugin '%s'", mn, pn)
		}
		if err = ValidatePluginType(pc["type"].(string), mn); err != nil {
			return nil, err
		}
		var rmi di.PluginPtr  // ret map item
		rmi = &di.Plugin {
			Module: mod,
			Type: pc["type"].(string), 
		} 
		rmi.Config, err = mod.Hooks.LoadConfigHook(pc, true, fmt.Sprintf("%s.%s", path, pn))
		if (err != nil) {
			return nil, err
		}
		ret[pn] = rmi

	}
	return ret, nil
}
