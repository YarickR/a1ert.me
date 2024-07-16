package dm_core

import (
	"dagproc/internal/di"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

/*
  "channels":
    {
      "root": {
        "plugins": [ "redis_in" ]
      },
      "warn_all": {
        "plugins": [ "core", "xmpp_out" ],
        "core": { "rules": [ {  "id": 0, "src": "root", "cond": "eq .labels.severity 'warning'" } ] },
        "xmpp_out": { "group": "Warnings" }
      },

*/

func parseChannel_getPluginList(chM map[string]interface{}, path string) ([]string, error) {
	var (
		ret []string
		err error
		pn interface{}
		pl []interface{}
	)

	ret = nil
 	err = di.ValidateConfig(`{"plugins!": [ "string" ]}`, chM, path)
	if err != nil {
		return nil, err
	}
	pl  = chM["plugins"].([]interface{}) 
	ret = make([]string, 0, len(pl))
	for _, pn = range pl {
		switch pn := pn.(type) {
		case string:
			ret = append(ret, pn)
		default:
			return nil, errors.New("wrong type for plugin")
		}
	}
	return ret, nil
}
func parseChannel(chConfig interface{}, chName string, path string) (di.ChannelPtr, error) {
	var ret di.ChannelPtr
	var err error
	var ok bool
	var chM map[string]interface{} //channel map
	var pl []string                // plugin list
	var pn, np string                  // plugin name, new path
	var pc interface{}             //plugin config

	var chT string = `
		{
        "plugins!": [ "string" ],
        "core": { "rules": [ { "id": 0, "src!": "string", "cond": "string" } ] } }
    }
	`
	np = fmt.Sprintf("%s.%s", path, chName)
	err = di.ValidateConfig(chT, chConfig, np)
	if err != nil {
		return nil, err
	}
	chM = chConfig.(di.MSI)
	pl, err = parseChannel_getPluginList(chM, np) // to keep parseChannel from being too bloated
	if err != nil {
		return nil, fmt.Errorf("channel '%s' %w", chName, err)
	}

	ret = &di.Channel{Name: chName}
	mLog.Debug().Str("channel", chName).Msg("Loading config")

	for _, pn = range pl {
		var pCtx di.ChannelPluginCtx
		pCtx.Plugin, ok = di.GCfg.Plugins[pn]
		if !ok {
			return nil, fmt.Errorf("unknown plugin '%s'", pn)
		}
		pc, ok = chM[pn]
		err = nil
		if ok {
			mLog.Debug().Str("channel", chName).Str("plugin", pn).Msg("Loading config")
			np = fmt.Sprintf("%s.%s.%s", path, chName, pn)
			pCtx.Config, err = pCtx.Plugin.Module.Hooks.LoadConfigHook(pc, false, np)
		} else {
			mLog.Debug().Str("channel", chName).Str("plugin", pn).Msg("Plugin %s has no channel-specific config")
		}
		if err == nil {
			if (pCtx.Plugin.Type & di.PT_IN) != 0 {
				ret.InPlugs = append(ret.InPlugs, pCtx)
			}
			if (pCtx.Plugin.Type & di.PT_OUT) != 0 {
				ret.OutPlugs = append(ret.OutPlugs, pCtx)
			}
			if (pCtx.Plugin.Type & di.PT_PROC) != 0 {
				ret.ProcPlugs = append(ret.ProcPlugs, pCtx)
			}
		} else {
			mLog.Error().Str("channel", chName).Str("plugin", pn).Err(err)
		}
	}
	return ret, err
}

func LoadChannelsConfig(jsc map[string]interface{}, path string) (map[string]di.ChannelPtr, error) {
	var ret map[string]di.ChannelPtr
	var err error
	mLog.Debug().Msg("LoadChannelsConfig")
	ret = make(map[string]di.ChannelPtr)
	err = nil
	// We have to do two passes - parse all channels and assign ids etc, and then bind sources and sinks
	for k, v := range jsc {
		var newC di.ChannelPtr
		newC, err = parseChannel(v, k, path)
		if err != nil {
			mLog.Error().Str("channel", k).Err(err).Msg("Error parsing channel")
			return nil, err
		}
		ret[k] = newC
	}
	mLog.Debug().Msgf("Loaded config: %v", ret)
	return ret, nil
}

func ChannelGetKeyValue(event di.Event, key string) interface{} {
	var err error
	var path []string = strings.Split(key, ".")

	if (len(path) <= 2) || (len(path[0]) > 0) {
		// Second condition is extraneous, it checks if  key starts with a dot. But let's stay safe
		log.Printf("Invalid key %s", key)
		return nil
	}
	var pathIdx int
	var cursor interface{}
	cursor = event
	for pathIdx = 1; pathIdx < len(path); pathIdx++ {
		var currPart string = path[pathIdx]
		if len(currPart) == 0 {
			log.Printf("Zero length part in key %s", key)
			return nil
		}
		switch cursor.(type) { // since this data is unmarshaled by json, we know we have a limited set of possible types
		case []interface{}: // array, path[pathIdx] should contain array index

			var idx int
			idx, err = strconv.Atoi(currPart)
			if (err != nil) || (idx < 0) || (idx >= len(cursor.([]interface{}))) {
				log.Printf("Invalid index %d, full key %s, event: %+v, cursor: %+v", idx, key, event, cursor)
				return nil
			}
			cursor = cursor.([]interface{})[idx]
		case map[string]interface{}: // map, path[pathIdx] should contain map key
			var ok bool
			if cursor, ok = cursor.(map[string]interface{})[currPart]; !ok {
				log.Printf("No %s found , full key %s, event: %+v, cursor: %+v", currPart, key, event, cursor)
				return nil
			}
		default:
			log.Printf("Non-indexable member %s found , full key %s, event: %+v, cursor: %+v", currPart, key, event, cursor)
			return nil
		}
	}
	return cursor
}
