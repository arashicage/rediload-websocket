package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/martini-contrib/render"

	"github.com/garyburd/redigo/redis"
	"github.com/satori/go.uuid"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2"

	"github.com/syndtr/goleveldb/leveldb"
)

func Scan(r render.Render, res http.ResponseWriter, req *http.Request) {

	data := make(map[string]interface{})

	data["title"] = "Redis 扫描"
	data["isScan"] = true

	data["log"] = uuid.NewV4().String()

	conf, err := ini.Load("conf/app.conf")
	if err != nil {
		fmt.Println("can not load file: conf/app.conf")
	}

	redis_urls := conf.Section("common").Key("redis_urls").String()

	data["redis_urls"] = redis_urls

	r.HTML(200, "scan", data)
}

func ScanWs(r render.Render, res http.ResponseWriter, req *http.Request) {

	req.ParseForm()

	url := req.FormValue("url")

	fmt.Println(url)

	ws, err := websocket.Upgrade(res, req, nil, 1024, 1024)
	defer ws.Close()
	if _, ok := err.(websocket.HandshakeError); ok {
		fmt.Println(res, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		fmt.Println("Cannot setup WebSocket connection:", err)
		return
	}

	cfg, err := ini.Load("conf/app.conf")
	if err != nil {
		fmt.Println("can not found file: conf/app.conf")
		return
	}

	if false {
		mongo_urls := cfg.Section("common").Key("mongo_urls").String()

		session, err := mgo.Dial(mongo_urls)
		if err != nil {
			panic(err)
		}
		defer session.Close()

		scanMulti(url, session, ws)
	}

	if true {
		l := cfg.Section("common").Key("leveldb").String()

		db, err := leveldb.OpenFile(l, nil)
		if err != nil {
			fmt.Println(err)
		}
		defer db.Close()

		scanMultiToLevelDB(url, db, ws)

	}

}

func scanMulti(url_redis string, session *mgo.Session, ws *websocket.Conn) {

	urls := strings.Split(url_redis, ",")

	for i, url := range urls {
		err := scan(strings.Trim(url, " "), session, ws, i)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
	}

}

type processBar struct {
	Type string `json:"type"`
	Msg  string `json:"msg"`
}

func scan(url_redis string, session *mgo.Session, ws *websocket.Conn, idx int) error {

	fmt.Println(url_redis, "----------------------------")

	info, err := QueryStats(url_redis) // 要加异常处理, 连接不同的时候...
	if err != nil {
		fmt.Println(err)
	}

	stat := FiltrateStats(info, "db0.keys")

	fmt.Println("当前实例键数量:", stat["db0.keys"])

	rs, err := redis.Dial("tcp", url_redis)
	defer rs.Close()
	if err != nil {
		fmt.Printf("%s\n", err)
		return err
	}

	next0 := "0"
	next1 := "0"
	next2 := "x"
	max := "10000"
	i := 0
	for {

		repl, err := redis.Values(rs.Do("SCAN", next1, "MATCH", "*", "COUNT", max))
		if err != nil {
			fmt.Printf("%s\n", err)
			return err
		}

		for _, val := range repl {

			switch val.(type) {
			case []uint8:

				next2, _ = redis.String(val, nil)
				// fmt.Println(next2)

			case []interface{}:
				keys, err := redis.Strings(val, nil)
				if err != nil {
					fmt.Printf("%s\n", err)
				}
				// for range keys {
				for _, key := range keys {
					// fmt.Println(key) //, "on redis", url)
					// github.com/go-mgo/mgo
					// 存到ssdb 上

					insertMongo(key, url_redis, session)
					i++

				}

			}
		}

		db0keys, err := strconv.ParseUint(stat["db0.keys"], 10, 64)
		pct := float64(i) / float64(db0keys) * 100
		fmt.Println("当前进度", pct)

		message := processBar{
			Type: "output",
			Msg:  strconv.Itoa(i) + "," + stat["db0.keys"] + "," + strconv.FormatFloat(pct, 'f', 2, 64) + "%" + "," + url_redis + "," + strconv.Itoa(idx),
		}

		ws.WriteJSON(message)

		if next2 == next0 {
			break
		} else {
			next1 = next2

			fmt.Println(next1)
		}

	}
	return err
}

type RedisRecord struct {
	KEY string
	URL string
}

func insertMongo(key, url string, session *mgo.Session) {

	session.SetMode(mgo.Monotonic, true)

	c := session.DB("redis").C("keys")
	err := c.Insert(&RedisRecord{KEY: key, URL: url})
	if err != nil {
		panic(err)
	}

	//	result := Person{}
	//	err = c.Find(bson.M{"name": "Ale"}).One(&result)
	//	if err != nil {
	//		panic(err)
	//	}

	//	fmt.Println("Phone:", result.Phone)

}

// ------------------------

func scanMultiToLevelDB(url_redis string, db *leveldb.DB, ws *websocket.Conn) {

	urls := strings.Split(url_redis, ",")

	for i, url := range urls {
		err := scanToLevelDB(strings.Trim(url, " "), db, ws, i)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
	}

}

func scanToLevelDB(url_redis string, db *leveldb.DB, ws *websocket.Conn, idx int) error {

	fmt.Println(url_redis, "----------------------------")

	info, err := QueryStats(url_redis) // 要加异常处理, 连接不同的时候...
	if err != nil {
		fmt.Println(err)
	}

	stat := FiltrateStats(info, "db0.keys")

	fmt.Println("当前实例键数量:", stat["db0.keys"])

	rs, err := redis.Dial("tcp", url_redis)
	defer rs.Close()
	if err != nil {
		fmt.Printf("%s\n", err)
		return err
	}

	next0 := "0"
	next1 := "0"
	next2 := "x"
	max := "10000"
	i := 0
	for {

		repl, err := redis.Values(rs.Do("SCAN", next1, "MATCH", "*", "COUNT", max))
		if err != nil {
			fmt.Printf("%s\n", err)
			return err
		}

		for _, val := range repl {

			switch val.(type) {
			case []uint8:

				next2, _ = redis.String(val, nil)
				// fmt.Println(next2)

			case []interface{}:
				keys, err := redis.Strings(val, nil)
				if err != nil {
					fmt.Printf("%s\n", err)
				}
				// for range keys {
				for _, key := range keys {
					// fmt.Println(key) //, "on redis", url)
					// github.com/go-mgo/mgo
					// 存到ssdb 上

					insertLevelDB(key, url_redis, db)
					i++

				}

			}
		}

		db0keys, err := strconv.ParseUint(stat["db0.keys"], 10, 64)
		pct := float64(i) / float64(db0keys) * 100
		fmt.Println("当前进度", pct)

		message := processBar{
			Type: "output",
			Msg:  strconv.Itoa(i) + "," + stat["db0.keys"] + "," + strconv.FormatFloat(pct, 'f', 2, 64) + "%" + "," + url_redis + "," + strconv.Itoa(idx),
		}

		ws.WriteJSON(message)

		if next2 == next0 {
			break
		} else {
			next1 = next2

			fmt.Println(next1)
		}

	}
	return err
}

func insertLevelDB(key, url string, db *leveldb.DB) {

	err := db.Put([]byte(key), []byte(url), nil)
	if err != nil {
		fmt.Println(err)
	}

}
