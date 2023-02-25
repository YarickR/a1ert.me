package main
import(
    "github.com/rs/zerolog/log"
    "dagproc/internal/dm_core"
	"github.com/gomodule/redigo/redis"
)

type MainConfig struct {
    ModList         string `redis:"modules"`
    IsShutdown      bool   `redis:"shutdown"`
}

func configConnect(configDSN string) (redis.Conn, error) {
	var err error
	var ret redis.Conn
	ret, err = redis.DialURL(configDSN)
	if (err == nil) {
        log.Debug().Str("DSN", configDSN).Msg("Connected to config DSN")
    }
    return ret, err
}
func loadMainConfig(rc redis.Conn) (MainConfig, error) {
	var err error
	var mCfg MainConfig
	reply, err := redis.Values(rc.Do("HGETALL", "settings"))
	if err == nil {
		err = redis.ScanStruct(reply, &mCfg)
	}
	return mCfg, err
}

func loadModuleConfig(rc redis.Conn, modName string) (dm_core.ModConfig, error) {
	var err error
	var mCfg dm_core.ModConfig
	log.Debug().Str("module", modName).Msg("Loading config from settings" + "_" + modName + " hash")
	mCfg, err = redis.Values(rc.Do("HGETALL", "settings" + "_" + modName))
	log.Debug().Str("module", modName).Msgf("Loaded config: %s", mCfg)
	return mCfg, err

}

