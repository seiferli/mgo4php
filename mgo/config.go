// Copyright © 2017 seiferli <469997798@qq.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mgo

import (
	"os"
	"fmt"
	"strconv"
)

type DbConfig struct {
	host     string
	port     string
	user     string
	passwd   string
	poolsize int
}

func (conf *DbConfig) initConfig(id int, filepath string) {
	conf.poolsize = 50
	if os.Getenv("ENV") == "pro" || os.Getenv("ENV") == "test" {
		conf.host = os.Getenv("MONGO_HOST")
		conf.port = os.Getenv("MONGO_PORT")
		conf.user = os.Getenv("MONGO_USER")
		conf.passwd = os.Getenv("MONGO_PASSWD")
		pool, err:= strconv.Atoi(os.Getenv("MONGO_POOLSIZE"))
		if err!= nil && pool>1 {
			conf.poolsize = pool
		}

	} else {
		ok, err := PathExists(filepath) //"config.ini"
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
	fmt.Println( ConsoleYellow("Connecting ")+ ConsoleRed(conf.user+ ":***@"+ conf.host + ":" + conf.port) )
	//mongodb://myuser:mypass@localhost:40001,otherhost:40001/mydb
	return conf.user + ":" + conf.passwd + "@" + conf.host + ":" + conf.port
}