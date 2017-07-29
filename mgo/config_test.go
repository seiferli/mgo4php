package mgo

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	conf := SetConfig("../config.ini")
	fmt.Println("INI_HOST: "+ conf.GetValue("MONGODB", "MONGO_HOST"))
	fmt.Println("INI_PORT: "+ conf.GetValue("MONGODB", "MONGO_PORT"))
	fmt.Println("INI_USER: "+ conf.GetValue("MONGODB", "MONGO_USER"))
	fmt.Println("INI_PASSWD: "+ conf.GetValue("MONGODB", "MONGO_PASSWD"))

	var db DbConfig
	db.initConfig(0)
	fmt.Println("RUL: "+ db.getDialString() )
}