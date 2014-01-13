package imageserver

import (
	"menteslibres.net/gosexy/redis"
)

const (
	POOLSIZE   = 200
	CACHE_TTL  = int64(3)
	REDIS_ADDR = "127.0.0.1"
	REDIS_PORT = 6379
)

type Conn struct {
	counter uint
	client  *redis.Client
}

var pool = make(chan *Conn, POOLSIZE)

func init() {
	go func() {
		for i := 0; i < POOLSIZE; i++ {
			var c = redis.New()
			if c.Connect(REDIS_ADDR, REDIS_PORT) == nil {
				pool <- &Conn{0, c}
			}
		}
	}()
}

func getCache(key string) (val []byte, err error) {
	var conn = getConn()
	var _val string
	_val, err = conn.client.Get(key)
	pool <- conn
	if err != nil {
		log.Error("cache get error %s", err)
	} else {
		val = []byte(_val)
	}
	return
}

func setCache(key string, val []byte) (bool, error) {
	var conn = getConn()

	var err error
	//str, err = conn.client.Set(key, val)
	_, err = conn.client.SetEx(key, CACHE_TTL, val)
	pool <- conn
	if err != nil {
		log.Error("cache set error %s", err)
	}
	return err != nil, err
}

func delCache(key string) (bool, error) {
	var conn = getConn()

	var err error
	//str, err = conn.client.Set(key, val)
	_, err = conn.client.Del(key)
	pool <- conn
	if err != nil {
		log.Error("cache set error %s", err)
	}
	return err != nil, err
}

func getConn() (conn *Conn) {
	conn = <-pool
	conn.counter++
	return
}
