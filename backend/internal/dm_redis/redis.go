package dm_redis
import (
	//"encoding/json"
	//"flag"
	"fmt"
	//"io/ioutil"
	//"os"
	"regexp"
	//"sync"
	//"text/template"
	//"time"

	//"github.com/gomodule/redigo/redis"
	"errors"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
    "dagproc/internal/di"
    "dagproc/internal/di_modplug"
)

var (
    mConfig RedisConfig
    mLog zerolog.Logger
)

func redisConfigKWDF_module (v interface{}, rcp RedisConfigPtr) error {
	switch t := v.(type) {
		case string:
			if (v.(string) != "redis") {
				return errors.New(fmt.Sprintf("module should be 'redis' instead of %s", v.(string)))
			}
			return nil
		default:
			return errors.New(fmt.Sprintf("Wrong type for 'module' keyword: %T, should be 'string'", t))
	}
}
func redisConfigKWDF_hooks (v interface{}, rcp RedisConfigPtr) error {
	switch t := v.(type) {
		case []string:
			return di_modplug.ValidateHooks(v.([]string), "redis")
		default:
			return errors.New(fmt.Sprintf("Wrong type for 'hooks' keyword: %T", t))
	}
}
func redisConfigKWDF_uri (v interface{}, rcp RedisConfigPtr) error {
	switch t := v.(type) {
		case string:
			var m bool
			if m, _ = regexp.Match("redis://.*:([0-9]+)/.*", v.([]byte)); m == true  { // host and port
				rcp.uri = v.(string)
				return nil
			}
			if m, _ = regexp.Match("redis:///.+", v.([]byte)); m == true { // unix domain socket
				rcp.uri = v.(string)
				return nil
			}
			return errors.New(fmt.Sprintf("%s does not look like redis uri", v.(string)))
		default:
			return errors.New(fmt.Sprintf("Wrong type for 'uri' keyword: %T, should be string 'redis://[user:password@]host:port/[db][?connection parameters]", t))
	}
}
func redisConfigKWDF_list (v interface{}, rcp RedisConfigPtr) error {
	switch t := v.(type) {
		case string:
			var m bool
			if m, _ = regexp.Match("[0-9a-zA-Z_]", v.([]byte)); m == true  {
				rcp.list = v.(string)
				return nil
			}
			return errors.New(fmt.Sprintf("%s does not look like list name", v.(string)))
		default:
			return errors.New(fmt.Sprintf("Wrong type for 'list' keyword: %T, should be string with proper redis identifier", t))
	}
}

func redisLoadConfig(config di.CFConfig) (di.PluginConfig, error) {
    var err error
    var ret RedisConfig
    var k string
    var v interface{}
    var kwdfm map[string]RedisConfigKWDF = map[string]RedisConfigKWDF {
    	"module":redisConfigKWDF_module,
    	"hooks":redisConfigKWDF_hooks,
    	"uri":redisConfigKWDF_uri,
    	"list":redisConfigKWDF_list,
    }
    for k, v = range config {
    	err = kwdfm[k](v, &ret)
    	if (err != nil) {
    		mLog.Error().Str("keyword", k).Err(err).Send()
    		return ret, err
    	}
    }
    return ret, nil
}
func ModInit() (di.ModHookTable, error) {
    mLog = log.With().Str("module", "redis").Logger()
    mLog.Debug().Msg("ModInit")
    return di.ModHookTable{ 
   		LoadConfigHook:      redisLoadConfig,
    	ReceiveEventHook:    nil,
    	SendEventHook:       nil,
    	ProcessEventHook:    nil,
    }, nil
}

/*
func redisLoadChannelDefs(rc redis.Conn, lastChannelId uint32) ([]*ChannelDef, uint32, error) {
	var err error
	var cdNum uint64
	var cdsList []string
	var channelDefs []*ChannelDef
	cdNum, err = redis.Uint64(rc.Do("LLEN", "channel_defs"))
	if uint32(cdNum) > lastChannelId {
		lastChannelId = uint32(cdNum)
	}
	channelDefs = make([]*ChannelDef, lastChannelId + 1, lastChannelId+1)
	cdsList, err = redis.Strings(rc.Do("LRANGE", "channel_defs", 0, cdNum))
	if err != nil {
		log.Printf("Error %s retrieving channel definitions from Redis\n", err)
		return nil, 0, err
	}
	for _, cdS := range cdsList {
		log.Printf("Channel def: %s\n", cdS)
		var chDef *ChannelDef
		chDef = new(ChannelDef)
		err = json.Unmarshal([]byte(cdS), chDef)
		if err != nil {
			log.Printf("Error %s parsing channel definition %s", err, cdS)
		} else {
			if ((channelDefs[chDef.Id] == nil) ||
				(channelDefs[chDef.Id].Id != chDef.Id) ||
				(channelDefs[chDef.Id].Version != chDef.Version) ||
				(chDef.Id == 0)) {
				err = channelParseRules(chDef)
				if (err == nil) {
					if (chDef.MsgFormat != "")  {
						chDef.MsgTemplate, err = template.New("").Parse(chDef.MsgFormat)
						if (err != nil) {
							log.Printf("Error %s parsing template %s", err, chDef.MsgFormat)
							chDef.MsgTemplate, err = template.New("").Parse("{{.}}")
						}
					} else
					if (chDef.Id == 0) {
						chDef.MsgTemplate, err = template.New("root").Parse("{{.}}")
					}
					channelDefs[chDef.Id] = chDef
				}
			}
			if (chDef.Id > lastChannelId) {
				lastChannelId = chDef.Id
			}
		}
	}
	channelPipeSrcsToSinks(channelDefs, lastChannelId) // doing it late so we can use forward channel ids as sources
	return channelDefs, lastChannelId, err
}
*/
/*

func redisWorker(config Config, alertChannel chan XmppMsg, wg *sync.WaitGroup) {
	log.Printf("Starting redis worker with redis at %s", config.RedisURI)
	var rc redis.Conn
	defer wg.Done()
	var err error
	rc, err = redis.DialURL(config.RedisURI)
	if err != nil {
		log.Printf("Error %s connecting to Redis at %s", err, config.RedisURI)
		return
	}
	defer rc.Close() // Safe to call close at this time, rc should be a valid connection handle
	var svcCfg ServiceConfig
	var channelDefs []*ChannelDef
	var lastChannelId uint32 = 0
	var channelDefVer uint32 = 0
	svcCfg, err = redisLoadServiceConfig(rc)
	if err != nil { // Error reporting should be done in the called function
		return
	}
	channelDefVer = svcCfg.ChannelDefVer
	channelDefs, lastChannelId, err = redisLoadChannelDefs(rc, svcCfg.LastChannelId)
	if err != nil {
		return
	}
	log.Printf("Last channel id: %d", lastChannelId)

	for svcCfg.IsShutdown == false {
		var reply []interface{}
		var channel, alert string
		reply, err = redis.Values(rc.Do("BLPOP", "alerts", 1))
		if err != nil {
			if err.Error() != "nil returned" {
				log.Printf("Error %s getting another alert", err)
			}
		} else {
			_, err = redis.Scan(reply, &channel, &alert)
			if err != nil {
				log.Printf("Error %s scanning reply", err)
				continue // looping early to save on indentation
			}
			var newSvcCfg ServiceConfig
			newSvcCfg, err = redisLoadServiceConfig(rc)
			if newSvcCfg.ChannelDefVer > channelDefVer {
				var newChannelDefs []*ChannelDef
				var newLastChannelId uint32
				newChannelDefs, newLastChannelId, err = redisLoadChannelDefs(rc, newSvcCfg.LastChannelId)
				if err != nil {
					log.Printf("Keeping previous config and channel definitions version")
				} else {
					svcCfg = newSvcCfg
					channelDefs = newChannelDefs
					lastChannelId = newLastChannelId
					channelDefVer = svcCfg.ChannelDefVer
					log.Printf("Loaded channel definitions version %d, last channel id %d, definitions:\n %+v\n", channelDefVer, lastChannelId, channelDefs)
				}
			}

			var parsedMsg AlertMsg
			err = json.Unmarshal([]byte(alert), &parsedMsg)
			if err != nil {
				log.Printf("***Error*** %s unmarshaling %s", err, alert)
			} else {
				var receivedTS uint64
				receivedTS = uint64(parsedMsg.ReceivedTS)
				log.Printf("Message received at %d\n", receivedTS)
				for _, parsedAlert := range parsedMsg.Message.Alerts {
					var groupsToDeliver map[string]string
					groupsToDeliver = make(map[string]string)
					parsedMsg.Matches = channelRunTheGauntlet(channelDefs, 0, 0, parsedAlert.(map[string]interface{}), groupsToDeliver, 0)
					if (len(groupsToDeliver) > 0) {
						var groupName, groupMsg string
						for groupName, groupMsg = range groupsToDeliver {
							log.Printf("'%s' <- %s", groupName, groupMsg)
							alertChannel <- XmppMsg{XmppGroup: groupName, Message: groupMsg }
						}
					} else {
						log.Printf("Undeliverable alert: %+v", parsedAlert)
						rc.Do("RPUSH", "undelivered", alert)
					}
					if (parsedMsg.Matches <= 1 ) {
						log.Printf("Message %+v didn't match anything", parsedAlert)
						rc.Do("RPUSH", "unmatched", alert)
					}
				}
			}
		}
	}

}

*/
