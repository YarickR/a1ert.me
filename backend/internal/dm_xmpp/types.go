package dm_xmpp
import (
    "sync"
)
type XmppConfigPtr *XmppConfig
type XmppConfig struct {
    server      string
    login       string
    password    string
    groupsURI   string
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
