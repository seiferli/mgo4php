package main

import (
	"api.pool.mongodb/mgo"
	rpc "github.com/hprose/hprose-golang/rpc/fasthttp"
	"github.com/valyala/fasthttp"
)

func main() {
	mgoResource := mgo.NewResource()
	service := rpc.NewFastHTTPService()
	//some methods
	service.AddFunction("changeResource", mgoResource.ChangeResource)
	service.AddFunction("getData", mgoResource.GetData)
	service.AddFunction("oneData", mgoResource.OneData)
	service.AddFunction("simpleInsert", mgoResource.SimpleInsert)
	service.AddFunction("simpleUpdate", mgoResource.SimpleUpdate)
	service.AddFunction("batchInsert", mgoResource.BatchInsert)
	service.AddFunction("deleteData", mgoResource.DeleteData)

	fasthttp.ListenAndServe(":8080", service.ServeFastHTTP)
}
