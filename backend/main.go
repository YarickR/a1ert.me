package main

import (
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
    modInitDispatcher map[string]dm_core.ModInitFunc
    modMap map[string]dm_core.Module
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
	var configDSN string
	var logLevel string
	var dryRun bool
	var _ls2llc map[string]zerolog.Level
	var config MainConfig
	var err error
	var moduleNames []string
	//var loadedModules map[string]dm_core.Module
	//var wg sync.WaitGroup
	var configConn redis.Conn

    _ls2llc = map[string]zerolog.Level{"debug": zerolog.DebugLevel, "info": zerolog.InfoLevel, "warn": zerolog.WarnLevel, "fatal": zerolog.FatalLevel}
    modInitDispatcher = map[string]dm_core.ModInitFunc{
        "core":     dm_core.ModInit,
        "xmpp":     dm_xmpp.ModInit,
        "redis":    dm_redis.ModInit,
    }

	flag.StringVar(&configDSN, "c", "", "config DSN (redis://<host[:port]>/[database id])")
	flag.StringVar(&logLevel, "l", "info", "Log level (debug|info|warn|fatal)")
	flag.BoolVar(&dryRun, "n", false, "Dry run (check config and exit)")
	flag.Parse()
	if configDSN == "" {
		log.Fatal().Msgf("Usage: %s -c <config DSN> [ -l <debug|info|warn|fatal>] [ -n for dry run]", os.Args[0])
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
    configConn, err = configConnect(configDSN)
    if (err != nil) {
        log.Fatal().Err(err).Msgf("Unable to load main config at %s", configDSN)
        os.Exit(1)
    }
    config, err = loadMainConfig(configConn)
    if (err == nil) {
        moduleNames = strings.Split(config.ModList, ",")
        moduleNames  = append([]string{"core"}, moduleNames...)
        modMap = make(map[string]dm_core.Module, len(moduleNames))
        var modName string
        for _, modName = range moduleNames {
           log.Debug().Str("module", modName).Msg("Starting init")
            if modInitFunc, ok := modInitDispatcher[modName]; ok {
                var mc dm_core.ModConfig
                mc, err = loadModuleConfig(configConn, modName)
                if (err == nil) {
                    var modDispTable dm_core.ModDispTable
                    modDispTable, err = modInitFunc()
                    if (err != nil) {
                        log.Fatal().Err(err).Str("module", modName).Msg("Unable to init module")
                        break
                    } else {
                        err = modDispTable.LoadConfig(mc)
                    }
                    if (err == nil) {
                        modMap[modName] = dm_core.Module{ Name: modName, Config: nil, DispTable: modDispTable }
                    } else {
                        log.Fatal().Err(err).Msg("Unable to parse loaded config")
                    }
                } else {
                    log.Fatal().Str("module", modName).Msg("Unable to load module config")
                }
            } else {
                err = errors.New(fmt.Sprintf("Unknown module %s", modName))
            }
            if (err != nil) {
                break
            }
        }
        if (err != nil) {
            os.Exit(3)
        }
    } else {
        log.Fatal().Err(err).Msgf("Cannot load main config from %s", configDSN)
        os.Exit(2)
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
