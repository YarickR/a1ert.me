package dm_xmpp

import (
	"sync"
    "dagproc/internal/di"
)

type XmppConfigPtr *XmppConfig
type XmppConfig struct {
    server      string
    login       string
    password    string
    groupsURI   string
    group       string
    template    di.TemplatePtr
}

type XmppMsg struct {
    XmppGroup   string
    Message     string
}

const JOIN_STATUS_NOT_JOINED = 0
const JOIN_STATUS_JOINING = 1
const JOIN_STATUS_JOINED = 2
type XmppGroup struct {
    JoinStatus    uint32
    GroupName     string
    GroupServer   string
    GroupNick     string
    Deferred    []string
    GroupLock  sync.Mutex
}
type XmppConfigKWDF func (v interface{}, xcp XmppConfigPtr) error 
type XmppConfigKWD struct  {
    dispFunc    XmppConfigKWDF
    dispFlags   uint
}