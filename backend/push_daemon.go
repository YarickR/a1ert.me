package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sync"
	"time"
	"text/template"
)

func redisLoadServiceConfig(rc redis.Conn) (ServiceConfig, error) {
	var err error
	var svcCfg ServiceConfig

	reply, err := redis.Values(rc.Do("HGETALL", "settings"))
	if err != nil {
		log.Printf("Error %s getting settings from Redis", err)
	} else {
		err = redis.ScanStruct(reply, &svcCfg)
		if err != nil {
			log.Printf("Error %s parsing settings in Redis", err)
		} else {
			log.Printf("Settings in Redis: %+v\n", svcCfg)
		}
	}
	return svcCfg, err
}

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
					var chRule ChannelRule
					for _, chRule = range chDef.Rules {
						if (chRule.Root.function != nil) { // Rule was successfully parsed
							channelAddSinkId(channelDefs[chRule.SrcChId], chDef.Id) // almost placeholder at this point 
																				  // will check for acyclicity later
						}
					}
				}
			}
			if chDef.Id > lastChannelId {
				lastChannelId = chDef.Id
			}
		}
	}
	return channelDefs, lastChannelId, err
}

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
				for _, alert := range parsedMsg.Message.Alerts {
					var groupsToDeliver map[string]string
					groupsToDeliver = make(map[string]string)
					parsedMsg.Matches = channelRunTheGauntlet(channelDefs, 0, 0, alert.(map[string]interface{}), groupsToDeliver, 0)
					if (len(groupsToDeliver) > 0) {
						var groupName, groupMsg string
						for groupName, groupMsg = range groupsToDeliver {
							log.Printf("'%s' <- %s", groupName, groupMsg)
							alertChannel <- XmppMsg{XmppGroup: groupName, Message: groupMsg }
						}
					} else {
						log.Printf("Undeliverable alert: %+v", alert)
						rc.Do("RPUSH", "undelivered", alert)
					}
					if (parsedMsg.Matches <= 1 ) {
						log.Printf("Message %+v didn't match anything", alert)
						rc.Do("RPUSH", "unmatched", alert)
					}
				}
			}
		}
	}

}

func xmppHandleMessage(c xmpp.Sender, p stanza.Packet) {
	var ok bool
	msg, ok := p.(stanza.Message)
	if ok {
		log.Printf("New message received: %+v\n", msg)
	} else {
		log.Printf("Error getting new message from packet %+v", p)
	}
	return
}

func xmppHandleIQ(c xmpp.Sender, p stanza.Packet) {
	log.Printf("New Info/Query received: %+v\n", p)
	return
}

func xmppHandlePresence(c xmpp.Sender, p stanza.Packet, jidLogin string, groupMap map[string]*XmppGroup) {
	var pres stanza.Presence
	var ok bool
	var err error
	var jidRe *regexp.Regexp
	var fromJid, myJid []string
	if pres, ok = p.(stanza.Presence); ok {
		log.Printf("Presence received; from: %s to: %s", pres.Attrs.From, pres.Attrs.To)
		jidRe, _ = regexp.Compile("^([^@]+)@([^/]+)/?(.*)$")
		myJid = jidRe.FindStringSubmatch(jidLogin)
		if (pres.Attrs.To == jidLogin) {
			fromJid = jidRe.FindStringSubmatch(pres.Attrs.From)
			if ((fromJid != nil) && (len(fromJid) >= 2) && (fromJid[3] == myJid[3])) {
				var cg *XmppGroup // current group
				if cg, ok = groupMap[fromJid[1]]; ok {
					log.Printf("Looks like we joined \"%s\" group, it has %d deferred messages, sending them", fromJid[1], len(cg.Deferred))
					for {
						var mq []string // message queue
						cg.GroupLock.Lock()

						if len(cg.Deferred) > 0 {
							mq = cg.Deferred
							cg.Deferred = make([]string, 0, 10)
							log.Printf("Still %d deferred messages for \"%s\"", len(mq), fromJid[1])
							// we're making list of alerts to send private to this thread , and empty it for others
							// so after we're done sending them we could check if there's more to send or we're good to finally
							// declare group status as properly joined and we could send alerts synchronously
						} else {
							cg.JoinStatus = JOIN_STATUS_JOINED
						}
						cg.GroupLock.Unlock()
						if JOIN_STATUS_JOINED == cg.JoinStatus {
							break
						} else {
							var alert string

							for _, alert = range mq {
								if len(alert) == 0 {
									log.Printf("Empty alert ?")
									continue
								}
								log.Printf("Going to send another alert: %s", alert)
								var msg stanza.Message
								msg = stanza.Message{
									Attrs: stanza.Attrs{
										Type: stanza.MessageTypeGroupchat,
										From: pres.Attrs.To,
										To:   fmt.Sprintf("%s@%s", cg.GroupName, cg.GroupServer),
									},
									Body: alert,
								}
								if err = c.Send(msg); err != nil {
									log.Printf("Error %s sending new message %+v", err, msg)
								}
							}
						}
					}
				}
			}
		}

	} else {
		log.Printf("Error getting presense from packet %+v", p)
	}
	return
}

func xmppHandleClientError(err error) {
	log.Printf("Client error %s\n", err)
}

// Shamelessly ripped from go-xmpp/cmd/fluuxmpp/xmppmuc.go
func joinMUC(c xmpp.Sender, toJID *stanza.Jid) error {
	return c.Send(stanza.Presence{Attrs: stanza.Attrs{To: toJID.Full()},
		Extensions: []stanza.PresExtension{
			stanza.MucPresence{
				History: stanza.History{MaxStanzas: stanza.NewNullableInt(0)},
			}},
	})
}

// Shamelessly ripped from go-xmpp/cmd/fluuxmpp/xmppmuc.go
func leaveMUCs(c xmpp.Sender, mucsToLeave []*stanza.Jid) {
	for _, muc := range mucsToLeave {
		if err := c.Send(stanza.Presence{Attrs: stanza.Attrs{
			To:   muc.Full(),
			Type: stanza.PresenceTypeUnavailable,
		}}); err != nil {
			log.Printf("error %+v leaving muc: %s", muc, err)
		}
	}
}
func xmppJoinDeferred(c xmpp.Sender, groupMap map[string]*XmppGroup) {
	var err error
	var cg *XmppGroup
	for _, cg = range groupMap {
		var gJID *stanza.Jid
		var gjs string
		gjs = fmt.Sprintf("%s@%s/%s", cg.GroupName, cg.GroupServer, cg.GroupNick)
		gJID, err = stanza.NewJid(gjs)
		if err != nil {
			log.Printf("Error %s creating group JID %s", err, gjs)
		} else {
			log.Printf("Starting to join %s", gjs)
			if err = joinMUC(c, gJID); err != nil {
				log.Printf("Error %s joining a group", err)
			} else {
				cg.GroupLock.Lock()
				cg.JoinStatus = JOIN_STATUS_JOINING
				cg.GroupLock.Unlock()
			}
		}
	}
}
func xmppWorker(config Config, alertChannel chan XmppMsg, wg *sync.WaitGroup) {
	var xmppCfg xmpp.Config
	var router *xmpp.Router
	var cm *xmpp.StreamManager
	var cli *xmpp.Client
	var err error
	var ConnStartTS, ConnOkTS int64
	var xmppGroups map[string]*XmppGroup
	var jidLogin string
	xmppGroups = make(map[string]*XmppGroup)

	log.Printf("Starting XMPP worker, server: %s, login: %s, groups: %s", config.XmppServer, config.JidLogin, config.GroupsURI)
	defer wg.Done()
	var _xmppServer string = config.XmppServer
	var _hasPort bool
	_hasPort, err = regexp.Match(":[0-9]+$", []byte(_xmppServer))
	if !_hasPort {
		_xmppServer = fmt.Sprintf("%s:5222", config.XmppServer)
	}
	jidLogin = fmt.Sprintf("%s@%s/%s", config.JidLogin, config.XmppServer, config.JidNick)
	xmppCfg = xmpp.Config{
		TransportConfiguration: xmpp.TransportConfiguration{
			Address: _xmppServer,
		},
		Jid:          jidLogin,
		Credential:   xmpp.Password(config.JidPwd),
		StreamLogger: os.Stderr,
		Insecure:     true,
		// TLSConfig: tls.Config{InsecureSkipVerify: true},
	}

	router = xmpp.NewRouter()
	router.HandleFunc("message", xmppHandleMessage)
	router.HandleFunc("iq", xmppHandleIQ)
	router.HandleFunc("presence", func(c xmpp.Sender, p stanza.Packet) {
		xmppHandlePresence(c, p, jidLogin, xmppGroups)
	})

	cli, err = xmpp.NewClient(&xmppCfg, router, xmppHandleClientError)
	if err != nil {
		log.Printf("Error %s creating new XMPP client, check your config", err)
		return
	}

	ConnStartTS = time.Now().Unix()
	ConnOkTS = 0		


	cm = xmpp.NewStreamManager(cli, func(c xmpp.Sender) {
		ConnOkTS = time.Now().Unix()
		xmppJoinDeferred(c, xmppGroups)
	})

	go func(sTmA *xmpp.StreamManager, wAgR *sync.WaitGroup) {
		err = sTmA.Run()
		if err != nil {
			log.Printf("Error %s running stream manager", err)
		}
		wAgR.Done()
	}(cm, wg)
	var alertMsg XmppMsg
	for alertMsg = range alertChannel {
//		log.Printf("Waiting for alert")
//		alertMsg = <-alertChannel
		log.Printf("Got new alert")
		var groupName string
		var cg *XmppGroup // current group

		groupName = alertMsg.XmppGroup

		var ok bool
		if _, ok = xmppGroups[groupName]; !ok {
			cg = new(XmppGroup)
			cg.JoinStatus = JOIN_STATUS_NOT_JOINED
			cg.Deferred = make([]string, 0, 10) // Arbitrary capacity number
			cg.GroupName = groupName
			cg.GroupServer = config.GroupsURI
			cg.GroupNick = config.JidNick
			xmppGroups[groupName] = cg
		}
		cg = xmppGroups[groupName]
		switch cg.JoinStatus {
			case JOIN_STATUS_NOT_JOINED:
				// Begin joining, queue new message
				if ConnOkTS == 0 {
					log.Printf("Not connected yet after %d seconds , deferring join and send attempts", (time.Now().Unix() - ConnStartTS))
					cg.GroupLock.Lock() // Mutex is not needed here , but for the sake of consistency
					cg.Deferred = append(cg.Deferred, alertMsg.Message)
					cg.GroupLock.Unlock()
				} else {
					var gJID *stanza.Jid
					gJID, err = stanza.NewJid(fmt.Sprintf("%s@%s/%s", cg.GroupName, cg.GroupServer, cg.GroupNick))
					if err != nil {
						log.Printf("Error %s creating group %s@%s", err, cg.GroupName, cg.GroupServer)
					} else {
						log.Printf("Starting to join %s@%s/%s", cg.GroupName, cg.GroupServer, cg.GroupNick)
						cg.GroupLock.Lock()
						cg.JoinStatus = JOIN_STATUS_JOINING
						if err = joinMUC(cli, gJID); err != nil {
							log.Printf("Error %s joining a group", err)
							cg.JoinStatus = JOIN_STATUS_NOT_JOINED
						} else {
							cg.Deferred = append(cg.Deferred, alertMsg.Message)
						}
						cg.GroupLock.Unlock()
					}
				}
			case JOIN_STATUS_JOINING:
				log.Printf("Still joining %s@%s", cg.GroupName, cg.GroupServer)
				cg.GroupLock.Lock()
				cg.Deferred = append(cg.Deferred, alertMsg.Message)
				cg.GroupLock.Unlock()
				// Still joining
			case JOIN_STATUS_JOINED:
				// Just post the message
				// At this point deferred message queue is empty , and we can send synchronously
				var msg stanza.Message
				msg = stanza.Message{
					Attrs: stanza.Attrs{
						Type: stanza.MessageTypeGroupchat,
						From: config.JidLogin,
						To:   fmt.Sprintf("%s@%s", cg.GroupName, cg.GroupServer),
					},
					Body: alertMsg.Message,
				}
				if err = cli.Send(msg); err != nil {
					log.Printf("Error %s sending new message %+v", err, msg)
				}
			}
	}
	cli.Disconnect()

}

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

func main() {
	var configFile string
	var config Config
	var err error
	var wg sync.WaitGroup
	var alertChannel chan XmppMsg
	flag.StringVar(&configFile, "c", "", "config file")
	flag.Parse()
	if configFile == "" {
		log.Printf("Usage: %s -c <config file>", os.Args[0])
		os.Exit(1)
	}
	config, err = loadConfig(configFile)
	if err != nil { // Actual error description is in the loadConfig func
		os.Exit(2)
	}
	alertChannel = make(chan XmppMsg, 1)
	wg.Add(3)
	go redisWorker(config, alertChannel, &wg)
	go xmppWorker(config, alertChannel, &wg)
	wg.Wait()
}
