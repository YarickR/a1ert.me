package dm_xmpp
import (
    "sync"
)

type XmppConfig struct {

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
