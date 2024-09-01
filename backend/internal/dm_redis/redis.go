package dm_redis

import (
	//"encoding/json"
	//"flag"
	//"io/ioutil"
	//"os"
	//"sync"
	//"text/template"
	//"time"

	//"github.com/gomodule/redigo/redis"
	"dagproc/internal/di"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	mConfig RedisConfig
	mLog    zerolog.Logger
)

func ModInit() (di.ModHookTable, error) {
	mLog = log.With().Str("module", "redis").Caller().Logger()
	mLog.Debug().Msg("ModInit")
	return di.ModHookTable{
		LoadConfigHook:	  redisLoadConfig,
		ReceiveMsgHook:	  redisRecvMsg,
		SendMsgHook:	  redisSendMsg,
		ProcessMsgHook:   nil,
	}, nil
}

func redisRecvMsg(chplct di.ChanPlugCtxPtr) (di.DagMsgPtr, error)  {
	var (
		ret error
		dams di.DagMsgPtr
    )
    dams = &di.DagMsg{ Data: nil, Channel: nil }
	return dams, ret
}

func redisSendMsg(dams di.DagMsgPtr, chplct di.ChanPlugCtxPtr) error {
	var (
		ret error
	)
	return ret
}
/*
func redisLoadChannelDefs(r                                                             c redis.Conn, lastChannelId uint32) ([]*ChannelDef, uint32, error) {
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
