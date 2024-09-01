package dm_http

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
	mConfig HttpConfig
	mLog    zerolog.Logger
)

func ModInit() (di.ModHookTable, error) {
	mLog = log.With().Str("module", "http").Caller().Logger()
	mLog.Debug().Msg("ModInit")
	return di.ModHookTable{
		LoadConfigHook:   	httpLoadConfig,
		ReceiveMsgHook: 	httpRecvMsg,
		SendMsgHook:    	httpSendMsg,
		ProcessMsgHook: 	httpProcessMsg,
	}, nil
}

func httpRecvMsg(chplct di.ChanPlugCtxPtr) (di.DagMsgPtr, error)  {
	var (
		ret error
		dams di.DagMsgPtr
    )
    dams = &di.DagMsg{ Data: nil, Channel: nil }
	return dams, ret
}
func httpSendMsg(dams di.DagMsgPtr, chplct di.ChanPlugCtxPtr) error {
	var (
		ret error
	)
	return ret
}

func httpProcessMsg(dams di.DagMsgPtr, chplct di.ChanPlugCtxPtr) (di.DagMsgPtr, error) {
	var (
		ret error
		odams di.DagMsgPtr
    )
    odams = &di.DagMsg{ Data: dams.Data, Channel: dams.Channel }
	return odams, ret
}
