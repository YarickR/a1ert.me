package "main"
import(
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gomodule/redigo/redis"
)


func configConnect(configDSN string) (redis.Conn, error) {
	var err error
	var ret redis.Conn
	ret, err = redis.DialURL(configDSN)
	if err != nil
}
func configLoadMainConfig(rc redis.Conn) (ServiceConfig, error) {
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

func configLoadPluginConfig(rc redis.Conn)
