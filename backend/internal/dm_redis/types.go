package dm_redis
import (
	"github.com/gomodule/redigo/redis"
)
type RedisConfigKWDF func (v interface{}, rcp RedisConfigPtr) error // kw == keyword, df == dispatcher func
type RedisConfigKWD struct {
	dispFunc 	RedisConfigKWDF
	dispFlags uint
}
type RedisConfigPtr *RedisConfig
type RedisConfig struct {
	uri string
	list string
}


type RedisCPC struct { // ChannelPluginContext
	conn redis.Conn
}