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
	service.AddFunction("all", mgoResource.AllData)
	service.AddFunction("one", mgoResource.OneData)
	service.AddFunction("count", mgoResource.CountData)

	service.AddFunction("insert", mgoResource.SimpleInsert)
	service.AddFunction("delete", mgoResource.DeleteData)
	service.AddFunction("update", mgoResource.SimpleUpdate)
	service.AddFunction("batchInsert", mgoResource.BatchInsert)

	fasthttp.ListenAndServe(":8080", service.ServeFastHTTP)
}
