package main

import (
	"dagproc/internal/di"      // di is dagproc internal
	"dagproc/internal/dm_core" // dm stands for dagproc module
	"dagproc/internal/dm_http"
	"dagproc/internal/dm_redis"
	"dagproc/internal/dm_xmpp"
	"flag"
	"fmt"
	"os"
	"time"

	//    "strings"
	//    "errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	var confFilePath string
	var logLevel string
	var dryRun bool
	var _ls2llc map[string]zerolog.Level
	var config di.GlobalConfig
	var err error
	var _mn string
	_ls2llc = map[string]zerolog.Level{"debug": zerolog.DebugLevel, "info": zerolog.InfoLevel, "warn": zerolog.WarnLevel, "fatal": zerolog.FatalLevel}

	di.ModHookMap = map[string]di.ModHooksFunc{
		"core":  dm_core.ModInit,
		"xmpp":  dm_xmpp.ModInit,
		"redis": dm_redis.ModInit,
		"http":  dm_http.ModInit,
	}

	flag.StringVar(&confFilePath, "c", "", "config file location")
	flag.StringVar(&logLevel, "l", "info", "Log level (debug|info|warn|fatal)")
	flag.BoolVar(&dryRun, "n", false, "Dry run (check config and exit)")
	flag.Parse()
	if confFilePath == "" {
		log.Fatal().Msgf("Usage: %s -c <config file> [ -l <debug|info|warn|fatal>] [ -n for dry run]", os.Args[0])
		os.Exit(1)
	}
	zerolog.TimestampFieldName = "t"
	zerolog.LevelFieldName = "l"
	zerolog.MessageFieldName = "m"
	zerolog.TimeFieldFormat = time.UnixDate
	//    log.Logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if trueLL, ok := _ls2llc[logLevel]; ok {
		zerolog.SetGlobalLevel(trueLL)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Info().Msgf("Unknown log level %s, actual log level set to info", logLevel)
	}
	di.ModMap = make(map[string]di.Module)
	for _mn = range di.ModHookMap {
		var _mhm di.ModHookTable
		_mhm, err = di.ModHookMap[_mn]()
		if err == nil {
			di.ModMap[_mn] = di.Module{
				Name:  _mn,
				Hooks: _mhm,
			}
		} else {
			break
		}
	}
	if err != nil {
		log.Error().Str("module", _mn).Msg("Error initializing module")
		os.Exit(1)
	}
	err = loadConfig(confFilePath)
	if err != nil {
		log.Error().Str("config path", confFilePath).Msg("Error loading main config")
		os.Exit(2)
	}
	log.Debug().Str("Loaded config", fmt.Sprintf("%v", config)).Send()
	/*
		config, err = loadConfig(configDSN)
		if err != nil { // Actual error description is in the loadConfig func
			os.Exit(2)
		}
		alertChannel = make(chan XmppMsg, 1)
		wg.Add(3)
		go redisWorker(config, alertChannel, &wg)
		go xmppWorker(config, alertChannel, &wg)
		wg.Wait()
	*/
}
