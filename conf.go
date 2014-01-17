package imageserver

import (
	"github.com/Unknwon/goconfig"
)

type Conf struct {
	config   *goconfig.ConfigFile
	ROOTDIR  string
	PIC404   string
	LOGFILE  string
	LOGLEVEL string

	POOLSIZE   int
	CACHE_TTL  int64
	REDIS_ADDR string
	L1_SIZE    int
}

var C *Conf = new(Conf)

func initConf() {
	config, err := goconfig.LoadConfigFile("conf.ini")
	if err != nil {
		panic(err)
	} else {
		C.config = config
		C.Load()
	}
}

func (conf *Conf) Load() {
	conf.ROOTDIR = conf.config.MustValue(goconfig.DEFAULT_SECTION, "rootdir", "/attachments/paopao")
	conf.PIC404 = conf.config.MustValue(goconfig.DEFAULT_SECTION, "pic404", "/static/default404.jpg")
	conf.LOGFILE = conf.config.MustValue(goconfig.DEFAULT_SECTION, "log", "imageserver.log")
	conf.LOGLEVEL = conf.config.MustValue(goconfig.DEFAULT_SECTION, "log_level", "warn")

	conf.POOLSIZE = conf.config.MustInt("cache", "pool_size", 200)
	conf.CACHE_TTL = conf.config.MustInt64("cache", "ttl", 3600)
	conf.REDIS_ADDR = conf.config.MustValue("cache", "redis_addr", "127.0.0.1:6379")
	conf.L1_SIZE = conf.config.MustInt("cache", "L1_size", 1024)
}
