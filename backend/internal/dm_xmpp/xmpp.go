package dm_xmpp

import (
	//	"encoding/json"
	//	"flag"
	//	"fmt"
	//	"io/ioutil"
	//	"os"
	//	"regexp"
	//	"sync"
	//	"text/template"
	//	"time"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	//	"gosrc.io/xmpp"
	//	"gosrc.io/xmpp/stanza"
	"dagproc/internal/di"
)

var (
	mConfig XmppConfig
	mLog    zerolog.Logger
)

func ModInit() (di.ModHookTable, error) {
	mLog = log.With().Str("module", "xmpp").Caller().Logger()
	mLog.Debug().Msg("ModInit")
	return di.ModHookTable{
		LoadConfigHook:   xmppLoadConfig,
		ReceiveEventHook: xmppReceive,
		SendEventHook:    xmppSend,
		ProcessEventHook: nil,
	}, nil
}

func xmppReceive() (di.Event, error) {
	var ret di.Event
	var err error
	err = nil
	return ret, err
}

func xmppSend(ev di.Event) error {
	var err error
	err = nil
	return err
}

/*
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
*/
