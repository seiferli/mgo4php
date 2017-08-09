package mgo

import (
	"api.pool.mongodb/logs"
	"encoding/json"
	"fmt"
	mgo "gopkg.in/mgo.v2"
	bson "gopkg.in/mgo.v2/bson"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
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
		res.logger.SetLogger(logs.AdapterConsole)

	} else {
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
	res.config.initConfig(0)
	session, err := mgo.Dial(res.config.getDialString())
	if err != nil {
		res.logger.Error(err.Error() + "[" + res.config.getDialString() + "]")
		//return err
	}
	res.resource = session
	res.resource.SetPoolLimit(res.config.poolsize)
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
		isDebug = true
	}
}

func (res *DbResource) processLog(log string) {
	if isDebug {
		res.logger.Info(log)
	}
}
func checkEssentialCondition(c map[string]string) apiFormat {
	if _, ok := c["database"]; !ok {
		return apiFormat{CODE_PARAMS, "Omit 'database' paramter when getData", nil}
	}
	if _, ok := c["collection"]; !ok {
		return apiFormat{CODE_PARAMS, "Omit 'collection' paramter when getData", nil}
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

func (res *DbResource) SimpleInsert(c map[string]string, d map[string]interface{}) string {
	format := checkEssentialCondition(c)
	if format.Code != CODE_SUCCESS {
		return format.Send()

	} else {
		instance := newInstance(res)
		defer instance.Close()

		collection := instance.DB(c["database"]).C(c["collection"])
		err := collection.Insert(handleInsertData(d))

		if err != nil {
			format.Code = CODE_DB
			format.Msg = string(err.Error())
		} else {
			res.processLog("[Insert]: '" + c["database"] + "." + c["collection"] +
				"' data: " + mapToString(d))
			format.Data = ""
		}
		return format.Send()
	}
}

func handleInsertData(d map[string]interface{}) bson.D {
	var final bson.D
	for dk, dv := range d {
		switch val := dv.(type) {
		case map[interface {}]interface {}:
			container := make(map[string]interface{})
			for sk, sv := range val {
				switch vval := sk.(type) {
				case string:
					container[vval]= sv
				}
			}
			final = append(final, bson.DocElem{dk, handleInsertData(container)})

		default:
			final = append(final, bson.DocElem{dk, val})
		}
	}
	return final
}

func (res *DbResource) BatchInsert(c map[string]string, d []map[string]interface{}) string {
	format := checkEssentialCondition(c)
	if format.Code != CODE_SUCCESS {
		return format.Send()

	} else {
		instance := newInstance(res)
		defer instance.Close()

		if len(d) > 1000 {
			Exception("Batch insert count must less than 1000.")

		} else {
			collection := instance.DB(c["database"]).C(c["collection"])

			var bsonD []interface{}
			for _, dv := range d {
				bsonD = append(bsonD, handleInsertData(dv))
			}

			fmt.Println(bsonD)

			err := collection.Insert(bsonD...)

			if err != nil {
				format.Code = CODE_DB
				format.Msg = string(err.Error())
			} else {
				res.processLog("[BatchInsert]: '" + c["database"] + "." + c["collection"] +
					"' count: " + strconv.Itoa(len(d)))
				format.Data = len(d)
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
			format.Msg = string(err.Error())
		} else {
			res.processLog("[Delete]: '" + c["database"] + "." + c["collection"] +
				"' Matched: " + strconv.Itoa(result.Matched) + "' Removed: " + strconv.Itoa(result.Removed))
			format.Data = result
		}
		return format.Send()
	}
}

func handleUpdateData(d map[interface{}]interface{}) bson.M {
	final := make(bson.M)
	if d[0] == "reflesh" {
		container := make(bson.M)
		for _, dv := range d {
			switch val := dv.(type) {
			case string:
			case map[interface{}]interface{}:
				for ck, cv := range val {
					switch sck := ck.(type) {
					case string:
						container[sck] = cv
					}
				}
			}
		}
		final["$set"] = container
	} else {
		for dk, dv := range d {
			switch val := dk.(type) {
			case string:
				final[val] = dv
			}
		}
	}
	return final
}

func (res *DbResource) SimpleUpdate(c map[string]string, w map[string]interface{}, d map[interface{}]interface{}) string {
	format := checkEssentialCondition(c)
	if format.Code != CODE_SUCCESS {
		return format.Send()

	} else {
		instance := newInstance(res)
		defer instance.Close()

		collection := instance.DB(c["database"]).C(c["collection"])
		data := handleUpdateData(d)
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

func (res *DbResource) ChangeResource(resourceId int) {
	res.config.initConfig(resourceId)
}

// invoke service end ==============
