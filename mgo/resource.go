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
	"github.com/seiferli/mgo4php/logs"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	mgo "gopkg.in/mgo.v2"
	bson "gopkg.in/mgo.v2/bson"
)

var initlock sync.Once
var isDebug bool

type DbResource struct {
	logger   *logs.BeeLogger
	config   DbConfig
	resource *mgo.Session
}

const (
	CODE_SUCCESS   = 0
	MSG_SUCCESS    = "ok"
	CODE_EXCEPTION = -1
	CODE_PARAMS    = -11 //
	CODE_DB        = -12 //eg：-12101
)

type apiFormat struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (format *apiFormat) Send() string {
	bt, err := json.MarshalIndent(format, "", "")
	if err != nil {
		panic("Have some error on apiFormat send")
	}
	return string(bt)
}

func Exception(msg string) {
	api := apiFormat{CODE_EXCEPTION, msg, ""}
	panic(api.Send())
}

func (res *DbResource) closeRes() {
	res.resource.Close()
}

func NewResource() *DbResource {
	lock := false
	initlock.Do(func() {
		lock = true
	})
	if !lock {
		panic("NewResource method can't run 2 times.")
	}

	res := new(DbResource)

	//init logger
	res.logger = logs.NewLogger(1000)
	if isDebug {
		fmt.Println( ConsoleYellow("Starting init CONSOLE logger")+ "...")
		res.logger.SetLogger(logs.AdapterConsole)

	} else {
		fmt.Println( ConsoleYellow("Starting init FILE logger")+ "...")
		//fmt.Println(runtime.GOOS, runtime.GOARCH, runtime.GOROOT(), os.Getenv("GOPATH"))
		// windows/darwin/linux/android/...   amd64   C:/goroot   F:/web/host/gopath
		logPath := runtime.GOROOT() + "/logs/"
		if os.Getenv("GOPATH") != "" {
			logPath = os.Getenv("GOPATH") + "/logs/"
		}
		err := os.MkdirAll(logPath, 0777)
		if err != nil {
			fmt.Printf("%s", err)
		}

		lgConf := `{"filename":"` + logPath + `api.mongodb.log"}`
		res.logger.SetLogger(logs.AdapterFile, lgConf)
	}

	//init mongodb db config
	fmt.Println( ConsoleYellow("Starting init mongodb config")+ "...")
	res.config.initConfig(0, "./config.ini")
	session, err := mgo.Dial(res.config.getDialString())
	if err != nil {
		res.logger.Error(err.Error() + "[" + res.config.getDialString() + "]")
		//return err
	}
	res.resource = session
	fmt.Println(ConsoleYellow("Setting pool size as " + ConvertToString(res.config.poolsize ))+ "...")
	res.resource.SetPoolLimit(res.config.poolsize)

	fmt.Println(ConsoleGreen("NewResource Complete !") )
	return res
}

func newInstance(res *DbResource) *mgo.Session {
	return res.resource.Copy()
}

func mapToString(m map[string]interface{}) string {
	var str string
	for k, v := range m {
		//switch val := v.(type) {
		//case int:
		//	str += "[int]" + k + ":" + strconv.Itoa(val) + ";"
		//case string:
		//	str += "[string]" + k + ":" + val + ";"
		//case byte:
		//	str += "[byte]" + k + ":" + string(val) + ";"
		//case bson.M:
		//	str += "[bson]" + k + ":" + mapToString(val)
		//default:
		js, err := json.Marshal(v)
		if err == nil {
			str += "[bson]" + k + ":" + string(js)
		} else {
			str += "[bson]" + k + ":" + "{}"
		}
		//}
	}
	if len(m) == 0 {
		str = "{}"
	}
	if !isDebug {
		str += "\n\r"
	}
	return str
}

func init() {
	if os.Getenv("ENV") != "pro" {
		isDebug = false  //开启日志写入性能会大幅下降
	}
}

func (res *DbResource) processLog(log string) {
	if isDebug {
		res.logger.Info(log)
	}
}
func checkEssentialCondition(c map[string]string) apiFormat {
	if _, ok := c["database"]; !ok {
		return apiFormat{CODE_PARAMS, "Omit 'database' paramter", nil}
	}
	if _, ok := c["collection"]; !ok {
		return apiFormat{CODE_PARAMS, "Omit 'collection' paramter", nil}
	}
	return apiFormat{CODE_SUCCESS, MSG_SUCCESS, ""}
}

func handleFieldAssetion(field interface{}) string {
	var filedName string
	switch val := field.(type) {
	case int:
		filedName = strconv.Itoa(val)
	case string:
		filedName = val
	case byte:
		filedName = string(val)
	}
	return filedName
}

func parseWhereRecursion(w map[interface{}]interface{}) bson.M {
	final := make(bson.M)
	if w[0] == "and" || w[0] == "or" {
		var bsonArr []bson.M
		for wk, wv := range w {
			switch val := wv.(type) {
			case []interface{}:
				container := make(map[interface{}]interface{})
				for ck, cv := range val {
					container[ck] = cv
				}
				if wk != 0 {
					bsonArr = append(bsonArr, parseWhereRecursion(container))
				}
			}
		}
		if len(bsonArr) < 2 {
			Exception("Must have two condition at least where using '$and' or '$or'.")
		}
		if w[0] == "and" {
			final["$and"] = bsonArr

		} else if w[0] == "or" {
			final["$or"] = bsonArr
		}

	} else if w[0] == "in" {
		final = bson.M{"$in": w[1]}

	} else if w[0] == ">" {
		final = bson.M{handleFieldAssetion(w[1]): bson.M{"$gt": w[2]}}

	} else if w[0] == ">=" {
		final = bson.M{handleFieldAssetion(w[1]): bson.M{"$gte": w[2]}}

	} else if w[0] == "<" {
		final = bson.M{handleFieldAssetion(w[1]): bson.M{"$lt": w[2]}}

	} else if w[0] == "<=" {
		final = bson.M{handleFieldAssetion(w[1]): bson.M{"$lte": w[2]}}

	} else if w[0] == "!" {
		final = bson.M{handleFieldAssetion(w[1]): bson.M{"$neq": w[2]}}

	} else if w[0] == "%" {
		reg := strings.Replace(handleFieldAssetion(w[2]), "/", "", -1)
		final = bson.M{handleFieldAssetion(w[1]): bson.M{"$regex": bson.RegEx{reg, "i"}}}

	} else if w[0] == "=" {
		final = bson.M{handleFieldAssetion(w[1]): bson.M{"$eq": w[2]}}

	} else {
		for wk, wv := range w {
			var instWk string
			instWk = handleFieldAssetion(wk)
			final[instWk] = wv
		}
	}
	//fmt.Println(final)
	return final
}

// invoke service start ==============
/**
type DbCondition struct {
    db  string        //database
    col string        //collection
    srt string 		  //sort field
    sel string        //select field
    off int64         //offset
    lmt int64         //limit
}
*/
func (res *DbResource) findData(instance *mgo.Session, c map[string]string, w map[interface{}]interface{}) *mgo.Query {
	collection := instance.DB(c["database"]).C(c["collection"])

	q := collection.Find(nil)
	var conditionLog string

	//some filter
	wbson := make(bson.M)
	wbson = parseWhereRecursion(w)

	if len(wbson) > 0 {
		//wbson= bson.M{"_id":bson.M{"$gt":1}}
		conditionLog += "Where:(" + mapToString(wbson) + "); "
		q = collection.Find(wbson)
	}

	if _, ok := c["offset"]; ok {
		offset, _ := strconv.Atoi(c["offset"])
		q.Skip(offset)
		conditionLog += "Offset:(" + strconv.Itoa(offset) + "); "
	}

	if _, ok := c["limit"]; ok {
		limit, _ := strconv.Atoi(c["limit"])
		q.Limit(limit)
		conditionLog += "Limit:(" + strconv.Itoa(limit) + "); "
	}

	if _, ok := c["sort"]; ok {
		sortField := strings.Split(c["sort"], ",")
		q.Sort(sortField...)
	}

	if _, ok := c["select"]; ok {
		selectField := strings.Split(c["select"], ",")
		sbson := make(bson.M)
		for _, v := range selectField {
			if v[0] == '-' {
				sbson[v[1:len(v)]] = 0
			} else if v[0] == '+' {
				sbson[v[1:len(v)]] = 1
			} else {
				sbson[v] = 1
			}
		}
		conditionLog += "Select:(" + mapToString(sbson) + "); "
		q.Select(sbson)
	}

	if conditionLog != "" {
		res.processLog("Condition: " + conditionLog)
	}
	return q
}

func (res *DbResource) AllData(c map[string]string, w map[interface{}]interface{}) string {
	format := checkEssentialCondition(c)
	if format.Code != CODE_SUCCESS {
		return format.Send()

	} else {
		instance := newInstance(res)
		defer instance.Close()

		q := res.findData(instance, c, w)
		var result []interface{}

		q.All(&result)
		res.processLog("[allData]: '" + c["database"] + "." + c["collection"] +
			"' rows count: " + strconv.Itoa(len(result)))
		format.Data = result
		return format.Send()
	}
}

func (res *DbResource) OneData(c map[string]string, w map[interface{}]interface{}) string {
	format := checkEssentialCondition(c)
	if format.Code != CODE_SUCCESS {
		return format.Send()

	} else {
		instance := newInstance(res)
		defer instance.Close()

		q := res.findData(instance, c, w)
		var result interface{}

		q.One(&result)
		res.processLog("[OneData]: '" + c["database"] + "." + c["collection"])
		format.Data = result
		return format.Send()
	}
}

func (res *DbResource) CountData(c map[string]string, w map[interface{}]interface{}) string {
	format := checkEssentialCondition(c)
	if format.Code != CODE_SUCCESS {
		return format.Send()

	} else {
		instance := newInstance(res)
		defer instance.Close()

		q := res.findData(instance, c, w)
		cnt, err := q.Count()
		if err != nil {
			format.Code = CODE_DB
			format.Msg = string(err.Error())

		} else {
			res.processLog("[Count]: '" + c["database"] + "." + c["collection"] +
				"' rows count: " + strconv.Itoa(cnt))
			format.Data = cnt
		}
		return format.Send()
	}
}

func (res *DbResource) SimpleInsert(c map[string]string, js string) string {
	format := checkEssentialCondition(c)
	if format.Code != CODE_SUCCESS {
		return format.Send()

	} else {
		instance := newInstance(res)
		defer instance.Close()

		collection := instance.DB(c["database"]).C(c["collection"])

		var data map[string]interface{}

		err := json.Unmarshal([]byte(js), &data )
		if err != nil {
			format.Code = CODE_DB
			format.Msg = "js "+ string(err.Error())+ " <"+ js + "> "

		} else {
			err = collection.Insert( data )
			if err != nil {
				format.Code = CODE_DB
				format.Msg = "insert "+ string(err.Error())
			} else {
				res.processLog("[Insert]: '" + c["database"] + "." + c["collection"] +
					"' data: " + js )
				format.Data = ""
			}
		}
		return format.Send()
	}
}

func (res *DbResource) BatchInsert(c map[string]string, js string ) string {
	format := checkEssentialCondition(c)
	if format.Code != CODE_SUCCESS {
		return format.Send()

	} else {
		instance := newInstance(res)
		defer instance.Close()

		var data []interface{}
		err := json.Unmarshal([]byte(js), &data )
		if err != nil {
			format.Code = CODE_DB
			format.Msg = "js "+ string(err.Error())+ " <"+ js + "> "

		} else {
			dataSize := len(data)
			if dataSize > 1000 {
				Exception("Batch insert count must less than 1000.")

			} else {
				collection := instance.DB(c["database"]).C(c["collection"])

				var output error
				for _, dd := range data {
					err = collection.Insert(dd)
					if err != nil {
						output= err
					}
				}

				if output != nil {
					format.Code = CODE_DB
					format.Msg = "bInsert "+ string(output.Error())
				} else {
					res.processLog("[BatchInsert]: '" + c["database"] + "." + c["collection"] +
						"' data: " + js )
					format.Data = dataSize
				}
			}
		}
		return format.Send()
	}
}

func (res *DbResource) DeleteData(c map[string]string, w map[string]interface{}) string {
	format := checkEssentialCondition(c)
	if format.Code != CODE_SUCCESS {
		return format.Send()

	} else {
		instance := newInstance(res)
		defer instance.Close()

		collection := instance.DB(c["database"]).C(c["collection"])
		result, err := collection.RemoveAll(w)

		res.processLog("Where:(" + mapToString(w) + "); ")

		if err != nil {
			format.Code = CODE_DB
			format.Msg = "delete "+ string(err.Error())
		} else {
			res.processLog("[Delete]: '" + c["database"] + "." + c["collection"] +
				"' Matched: " + strconv.Itoa(result.Matched) + "' Removed: " + strconv.Itoa(result.Removed))
			format.Data = result
		}
		return format.Send()
	}
}

/**
 * @deprecated this function already not used.
 *//*
func handleInsertData(d map[string]interface{}) bson.D {
	var final bson.D
	for dk, dv := range d {
		switch val := dv.(type) {
		case map[interface{}]interface{}:
			container := make(map[string]interface{})
			for sk, sv := range val {
				switch vval := sk.(type) {
				case string:
					container[vval] = sv
				}
			}
			final = append(final, bson.DocElem{dk, handleInsertData(container)})

		default:
			final = append(final, bson.DocElem{dk, val})
		}
	}
	return final
}
*/
/**
 * @deprecated this function already not used.
 *//*
func handleUpdateData(d map[interface{}]interface{}) bson.M {
	final := make(bson.M)
	if d[0] == "reflesh" {
		switch val:= d[1].(type) {
		case map[interface{}]interface{}:
			final["$set"] = handleUpdateData(val)
		}
	} else {
		for dk, dv := range d {
			switch val := dv.(type) {
			case map[interface{}]interface{}:
				switch vval := dk.(type) {
				case string:
					final[vval] = handleUpdateData(val)
				}
			default:
				switch vval := dk.(type) {
				case string:
					final[vval] = val
				}
			}
		}
	}
	return final
}
*/
/**
 * @deprecated this function already not used.
 *//*
func (res *DbResource) SimpleUpdate(c map[string]string, w map[string]interface{}, d map[interface{}]interface{}) string {
	format := checkEssentialCondition(c)
	if format.Code != CODE_SUCCESS {
		return format.Send()

	} else {
		instance := newInstance(res)
		defer instance.Close()

		collection := instance.DB(c["database"]).C(c["collection"])
		data := handleUpdateData(d)

		//handle object ID
		str, ok := w["_id_"]
		if ok {
			var oId string
			switch val := str.(type) {
			case int:
				oId = strconv.Itoa(val)
			case string:
				oId = val
			case byte:
				oId = string(val)
			}
			w["_id"]= bson.ObjectIdHex(oId)
			delete(w,"_id_")  //use _id_ replace into _id
		}

		err := collection.Update(w, data)

		res.processLog("Where:(" + mapToString(w) + "); ")

		if err != nil {
			format.Code = CODE_DB
			format.Msg = string(err.Error())
		} else {
			res.processLog("[Update]: '" + c["database"] + "." + c["collection"] +
				"' data: " + mapToString(data))
			format.Data = ""
		}

		return format.Send()
	}
}
*/
/**
 * mongodb command : "db.collection.update(criteria,objNew,upsert,multi)", upsert & multi is FLASE default.
 *    you can using $PARAM in "djs".
 * $inc ：increment the match field value.		eg: {$unset:{field:N}}
 * $set ：modify the match field value			eg: {$unset:{field:S}}
 * $unset ：delete the match field value			eg: {$unset:{field:1}}
 * $push ：add one element into the match field  				eg: {$push:{"field":"elementN"}}
 * $pushAll ：add element into the field value(more then one)	eg: {$pushAll:{"field":["e1","e2"]}}
 * $addToSet : like $push if the value not exist				eg: {$push:{"field":"Michael"}}
 * $pull ：delete the match field value							eg: {$pull:{"field":"elementN"}}
 * $pullAll ：delete the match field value(more then one)		eg: {$pullAll:{"field":["e1","e2"]}}
 * $pop : delete first element:-1, or last element:1			eg: {$pop:{"field":-1/1}}
 * $rename : change the field name								eg: {$rename:{"field1":"field2"}}
 * ...
 *
 */
func (res *DbResource) CombineUpdate(c map[string]string, wjs string, djs string, upsert bool) string {
	format := checkEssentialCondition(c)
	if format.Code != CODE_SUCCESS {
		return format.Send()

	} else {
		instance := newInstance(res)
		defer instance.Close()

		collection := instance.DB(c["database"]).C(c["collection"])

		var wdata, udata map[string]interface{}

		err1 := json.Unmarshal([]byte(wjs), &wdata )
		if err1 != nil {
			format.Code = CODE_DB
			format.Msg = "wjs "+ string(err1.Error())+ " <"+ wjs + "> "
		}
		err2 := json.Unmarshal([]byte(djs), &udata )
		if err2 != nil {
			format.Code = CODE_DB
			format.Msg = "djs "+ string(err2.Error())+ " <"+ djs + "> "
		}

		if err1==nil && err2==nil {
			var err error
			if upsert==true {
				result, errU := collection.Upsert(wdata, udata)
				if errU==nil {
					err = errU
				}
				res.processLog("Matched: " + strconv.Itoa(result.Matched) + "' Updated: " + strconv.Itoa(result.Updated))

			} else {
				err = collection.Update(wdata, udata)
			}

			if err != nil {
				format.Code = CODE_DB
				format.Msg = "update "+ string(err.Error())
			} else {
				res.processLog("[Update]: '" + c["database"] + "." + c["collection"] + "; Where:(" +
					wjs + "); Data:(" + djs + "); Upsert:" + strconv.FormatBool(upsert)  )
				format.Data = ""
			}
		}
		return format.Send()
	}
}

func (res *DbResource) ChangeResource(resourceId int) {
	res.config.initConfig(resourceId, "config.ini")
}

// invoke service end ==============
