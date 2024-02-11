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
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
    "dagproc/internal/di"
)

var (
    mConfig HttpConfig
    mLog zerolog.Logger
)

func ModInit() (di.ModHookTable, error) {
    mLog = log.With().Str("module", "http").Logger()
    mLog.Debug().Msg("ModInit")
    return di.ModHookTable{ 
   		LoadConfigHook:      httpLoadConfig,
    	ReceiveEventHook:    httpReceiveEvent,
    	SendEventHook:       httpSendEvent,
    	ProcessEventHook:    httpProcessEvent,
    }, nil
}

func httpReceiveEvent() (di.Event, error) {
	var ret di.Event
	var err error
	return ret, err
}
func httpSendEvent(ev di.Event) error {
	var err error
	return err
}
func httpProcessEvent(ev di.Event) error {
	var err error
	return err
}

