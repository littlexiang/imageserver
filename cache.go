package imageserver

import (
	//"menteslibres.net/gosexy/redis"
	"github.com/garyburd/redigo/redis"
	"github.com/golang/groupcache/lru"
	"time"
)

//var pool chan *Conn
var pool *redis.Pool
var L1 *lru.Cache

func initPool() {
	L1 = lru.New(C.L1_SIZE)

	pool = &redis.Pool{
		MaxIdle:     C.POOLSIZE,
		MaxActive:   0,
		IdleTimeout: 3600 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", C.REDIS_ADDR)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func getCache(key string) (val []byte, err error) {
	_val, ok := L1.Get(key)
	if ok {
		val = _val.([]byte)
		log.Fine("L1 hit %s", key)
		S.IncL1Hit()
	} else {
		log.Fine("L1 miss %s", key)
		S.IncL1Miss()

		var conn = pool.Get()
		defer conn.Close()
		val, err = redis.Bytes(conn.Do("GET", key))

		if err != nil {
			log.Fine("L2 miss %s %s", key, err)
			S.IncL2Miss()
		} else {
			S.IncL2Hit()
			L1.Add(key, val)
		}
	}

	return
}

func setCache(key string, val []byte) (bool, error) {
	L1.Add(key, val)

	var conn = pool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", key, string(val), "EX", C.CACHE_TTL)

	if err != nil {
		log.Error("L2 set error %s", err)
	}
	return err != nil, err
}

func delCache(key string) (err error) {
	L1.Remove(key)
	var conn = pool.Get()
	defer conn.Close()
	_, err = conn.Do("DEL", key)
	return
}
