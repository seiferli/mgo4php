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