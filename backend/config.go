package main

import (
	"dagproc/modplug"
	"io"
	"os"
	"github.com/rs/zerolog/log"
)

type GlobalConfig struct {
	Plugins			map[string]interface{} 
	Channels		map[string]interface{} 
	Templates		map[string]interface{} 
}

func loadConfig(confFilePath string) (GlobalConfig, error) {
	var err error
	var gCfg GlobalConfig
	var cfr io.Reader

	return gCfg, err
}
