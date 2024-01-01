package main

import (
	"dagproc/modplug"
	"dagproc/internal/dm_core" // dm stands for dagproc module
	"dagproc/internal/dm_redis"
	"dagproc/internal/dm_xmpp"
	"flag"
	"fmt"
	"os"
	"time"
    "strings"
    "errors"
	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog"
    "github.com/rs/zerolog/log"

)

var (
    modInfoMap 	map[string]modplug.ModInfoFunc // modInfoMap is populated manually early in main()
    modMap 		map[string]modplug.ModuleInfo  // modMap is populated by reading modInfoMap and calling ModInfoFunc's for each module
)


/*
func loadConfig(configFile string) (Config, error) {
	var fi os.FileInfo
	var err error
	var cfg Config
	fi, err = os.Stat(configFile)
	if err != nil {
		log.Printf("%s\n", err)
		return cfg, err
	}
	if !fi.IsDir() && (fi.Size() > 0) && (fi.Size() < 64*1024*1024) {
		var buf []byte
		buf, err = ioutil.ReadFile(configFile)
		if err != nil {
			log.Printf("%s\n", err)
			return cfg, err
		}

		err = json.Unmarshal(buf, &cfg)
		if err != nil {
			log.Printf("Error %s parsing config %s", err, buf)
			return cfg, err
		}
		if (cfg.JidLogin == "") || (cfg.JidPwd == "") || (cfg.RedisURI == "") || (cfg.GroupsURI == "") {
			err = fmt.Errorf("Incomplete config %s", configFile)
			log.Printf("%s\n", err)
			return cfg, err
		}
		return cfg, nil
	}
	return cfg, fmt.Errorf("Config file %s does not seem to be valid\n", configFile)

}
*/


func main() {
	var confFilePath string
	var logLevel string
	var dryRun bool
	var _ls2llc map[string]zerolog.Level
	var config GlobalConfig
	var err error
	var _mn string
	var _mhf modplug.ModHooksFunc
    _ls2llc = map[string]zerolog.Level{"debug": zerolog.DebugLevel, "info": zerolog.InfoLevel, "warn": zerolog.WarnLevel, "fatal": zerolog.FatalLevel}

    modInfoMap = map[string]modplug.ModInfoFunc {
        "core": 	dm_core.ModInit,
        "xmpp": 	dm_xmpp.ModInit,
        "redis": 	dm_redis.ModInit,
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
    log.Logger = log.With().Caller().Logger()
    if trueLL, ok := _ls2llc[logLevel]; ok {
        zerolog.SetGlobalLevel(trueLL)
    } else {
        zerolog.SetGlobalLevel(zerolog.InfoLevel)
        log.Info().Msgf("Unknown log level %s, actual log level set to info", logLevel)
    }
	modMap = make(map[string]modplug.ModuleInfo)
    for _mn, _mhf = range modInfoMap {
    	modMap[_mn] = modplug.ModuleInfo {
	    	Name: 	_mn,
	    	Config: make([]interface{}, 0),
	    	Hooks:  _mhf(),
    	}
    }

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
