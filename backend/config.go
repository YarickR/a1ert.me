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

func loadConfig(confFilePath string) error {
	var err error
	var gcj gcjf;

	var fi os.FileInfo
	var cfData []byte
	fi, err = os.Stat(confFilePath)
	if (err != nil) {
		log.Error().Err(err).Msgf("")
		return err
	}
	if (fi.IsDir() || fi.Size() > 1*1024*1024) {
		err = errors.New(fmt.Sprintf("%s - not a file or too large for a config file", confFilePath))
		log.Error().Str("size", fmt.Sprintf("%d", fi.Size())).Str("is_dir", fmt.Sprintf("%t", fi.IsDir())).Err(err).Send()
		return err
	} 
	cfData, err = os.ReadFile(confFilePath)
	if (err != nil) {
		log.Error().Err(err).Send()
		return err		
	}
	err = json.Unmarshal(cfData, &gcj)
	if (err != nil) {
		log.Error().Err(err).Send()
		return err		
	}
	// Templates go first, as plugins may reference them
	di.GCfg.Templates, err = dm_core.LoadTemplatesConfig(gcj.TmplDescr, "templates")
	if (err != nil) {
		log.Error().Msg("Error loading templates config")		
		return err
	}
	di.GCfg.Plugins, err = di_modplug.LoadPluginsConfig(gcj.PlugDescr, "plugins")
	if (err != nil) {
		log.Error().Err(err).Msg("Error loading plugins config")
		return err		
	}
	di.GCfg.Channels, err = dm_core.LoadChannelsConfig(gcj.ChanDescr, "channels")
	if (err != nil) {
		log.Error().Msg("Error loading channels config")
		return err		
	}
	return nil
}
