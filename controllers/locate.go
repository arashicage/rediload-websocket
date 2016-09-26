package controllers

import (
	// "log"
	"fmt"
	"net/http"
	"time"
	// "strconv"
	"github.com/gorilla/websocket"
	"github.com/martini-contrib/render"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/syndtr/goleveldb/leveldb"
)

func Locate(r render.Render, res http.ResponseWriter, req *http.Request) {
	data := make(map[string]interface{})

	r.HTML(200, "locate", data)
}

func LocateWsFromLevelDB(r render.Render, res http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	fplx := req.FormValue("fplx")
	fpdm := req.FormValue("fpdm")
	fphm := req.FormValue("fphm")
	key := fplx + ":" + fpdm + fphm

	cfg, err := ini.Load("conf/app.conf")
	if err != nil {
		fmt.Println("can not found file: conf/app.conf")
		return
	}

	l := cfg.Section("common").Key("leveldb").String()

	ws, err := websocket.Upgrade(res, req, nil, 1024, 1024)
	defer ws.Close()
	if _, ok := err.(websocket.HandshakeError); ok {
		fmt.Println(res, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		fmt.Println("Cannot setup WebSocket connection:", err)
		return
	}

	db, err := leveldb.OpenFile(l, nil)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	url, err := db.Get([]byte(key), nil)

	if err != nil {
		message := XXX{
			Type: "output",
			Msg: "<tr>" +
				"<td >" + time.Now().String()[11:19] + "</td>" +
				"<td >" + fplx + "</td>" +
				"<td >" + fpdm + "</td>" +
				"<td >" + fphm + "</td>" +
				"<td >" + "查不到" + "</td>" +
				"</tr>",
		}

		ws.WriteJSON(message)

		// panic(err)
	} else {
		message := XXX{
			Type: "output",
			Msg: "<tr>" +
				"<td >" + time.Now().String()[11:19] + "</td>" +
				"<td >" + fplx + "</td>" +
				"<td >" + fpdm + "</td>" +
				"<td >" + fphm + "</td>" +
				"<td >" + string(url) + "</td>" +
				"</tr>",
		}

		ws.WriteJSON(message)
	}

}

func LocateWs(r render.Render, res http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	fplx := req.FormValue("fplx")
	fpdm := req.FormValue("fpdm")
	fphm := req.FormValue("fphm")
	key := fplx + ":" + fpdm + fphm

	cfg, err := ini.Load("conf/app.conf")
	if err != nil {
		fmt.Println("can not found file: conf/app.conf")
		return
	}

	mongo_urls := cfg.Section("common").Key("mongo_urls").String()

	ws, err := websocket.Upgrade(res, req, nil, 1024, 1024)
	defer ws.Close()
	if _, ok := err.(websocket.HandshakeError); ok {
		fmt.Println(res, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		fmt.Println("Cannot setup WebSocket connection:", err)
		return
	}

	session, err := mgo.Dial(mongo_urls)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	c := session.DB("redis").C("keys")

	result := RedisRecord{}
	err = c.Find(bson.M{"key": key}).One(&result)
	if err != nil {
		message := XXX{
			Type: "output",
			Msg: "<tr>" +
				"<td >" + time.Now().String()[11:19] + "</td>" +
				"<td >" + fplx + "</td>" +
				"<td >" + fpdm + "</td>" +
				"<td >" + fphm + "</td>" +
				"<td >" + "查不到" + "</td>" +
				"</tr>",
		}

		ws.WriteJSON(message)

		// panic(err)
	} else {
		message := XXX{
			Type: "output",
			Msg: "<tr>" +
				"<td >" + time.Now().String()[11:19] + "</td>" +
				"<td >" + fplx + "</td>" +
				"<td >" + fpdm + "</td>" +
				"<td >" + fphm + "</td>" +
				"<td >" + result.URL + "</td>" +
				"</tr>",
		}

		ws.WriteJSON(message)
	}

}

type XXX struct {
	Type string `json:"type"`
	Msg  string `json:"msg"`
}
