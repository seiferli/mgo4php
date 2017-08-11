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

package main

import (
	rpc "github.com/hprose/hprose-golang/rpc/fasthttp"
	"github.com/seiferli/mgo4php/mgo"
	"github.com/valyala/fasthttp"
)

func main() {
	mgoResource := mgo.NewResource()
	service := rpc.NewFastHTTPService()
	//some methods
	//service.AddFunction("changeResource", mgoResource.ChangeResource)

	service.AddFunction("all", mgoResource.AllData)
	service.AddFunction("one", mgoResource.OneData)
	service.AddFunction("count", mgoResource.CountData)

	service.AddFunction("insert", mgoResource.SimpleInsert)
	service.AddFunction("delete", mgoResource.DeleteData)
	service.AddFunction("update", mgoResource.SimpleUpdate)
	service.AddFunction("batchInsert", mgoResource.BatchInsert)

	fasthttp.ListenAndServe(":8080", service.ServeFastHTTP)
}
