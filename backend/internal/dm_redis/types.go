package dm_redis

type RedisConfigKWDF func (v interface{}, rcp RedisConfigPtr) error // kw == keyword, df == dispatcher func

type RedisConfigPtr *RedisConfig
type RedisConfig struct {
	uri string
	list string
}
