package mgo

import (
	"os"
	"strconv"
)

type DbConfig struct {
	host     string
	port     string
	user     string
	passwd   string
	poolsize int
}

func (conf *DbConfig) initConfig(id int) {
	conf.poolsize = 50
	if os.Getenv("ENV") == "pro" || os.Getenv("ENV") == "test" {
		conf.host = os.Getenv("MONGO_HOST")
		conf.port = os.Getenv("MONGO_PORT")
		conf.user = os.Getenv("MONGO_USER")
		conf.passwd = os.Getenv("MONGO_PASSWD")

	} else {
		ok, err := PathExists("config.ini")
		if !ok {
			panic("Please copy <config.ini.sample> to <config.ini> at the same directory.")
		}

		ini := SetConfig("config.ini")  //dir with main.go
		conf.host = ini.GetValue("MONGODB", "MONGO_HOST")
		conf.port = ini.GetValue("MONGODB", "MONGO_PORT")
		conf.user = ini.GetValue("MONGODB", "MONGO_USER")
		conf.passwd = ini.GetValue("MONGODB", "MONGO_PASSWD")
		pool, err:= strconv.Atoi(ini.GetValue("MONGODB", "MONGO_POOLSIZE"))
		if err!= nil && pool>1 {
			conf.poolsize = pool
		}
	}
	//switch (id) {} //do some configuration
}

func (conf *DbConfig) getDialString() string {
	//mongodb://myuser:mypass@localhost:40001,otherhost:40001/mydb
	return conf.user + ":" + conf.passwd + "@" + conf.host + ":" + conf.port
}
