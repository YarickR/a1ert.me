package main

import (
//	"io"
	"os"
	"fmt"
	"errors"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"dagproc/internal/di_modplug"
	"dagproc/internal/di"
 	"dagproc/internal/dm_core"
)


type gcjf struct {
	PlugDescr       map[string]interface{} `json:"plugins"`
	ChanDescr       map[string]interface{} `json:"channels"`
	TmplDescr       map[string]interface{} `json:"templates"`
}

func loadConfig(confFilePath string, modMap map[string]di.Module) (di.GlobalConfig, error) {
	var err error
	var gCfg di.GlobalConfig
	var gcj gcjf;

	var fi os.FileInfo
	var cfData []byte
	fi, err = os.Stat(confFilePath)
	if (err != nil) {
		log.Error().Err(err).Msgf("")
		return gCfg, err
	}
	if (fi.IsDir() || fi.Size() > 1*1024*1024) {
		err = errors.New(fmt.Sprintf("%s - not a file or too large for a config file", confFilePath))
		log.Error().Str("size", fmt.Sprintf("%d", fi.Size())).Str("is_dir", fmt.Sprintf("%t", fi.IsDir())).Err(err).Send()
		return gCfg, err
	} 
	cfData, err = os.ReadFile(confFilePath)
	if (err != nil) {
		log.Error().Err(err).Send()
		return gCfg, err		
	}
	err = json.Unmarshal(cfData, &gcj)
	if (err != nil) {
		log.Error().Err(err).Send()
		return gCfg, err		
	}
	// Templates go first, as plugins may reference them
	gCfg.Templates, err = dm_core.LoadTemplatesConfig(gcj.TmplDescr)
	if (err != nil) {
		log.Error().Msg("Error loading templates config")		
		return gCfg, err
	}
	gCfg.Plugins, err = di_modplug.LoadPluginsConfig(gcj.PlugDescr, modMap)
	if (err != nil) {
		log.Error().Err(err).Msg("Error loading plugins config")
		return gCfg, err		
	}
	gCfg.Channels, err = dm_core.LoadChannelsConfig(gcj.ChanDescr, gCfg.Plugins)
	if (err != nil) {
		log.Error().Msg("Error loading channels config")
		return gCfg, err		
	}
	return gCfg, nil
}
