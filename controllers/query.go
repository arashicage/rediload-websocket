package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/garyburd/redigo/redis"
	"github.com/martini-contrib/render"

	"common/ini"
)

type fpxx struct {
	FPXX map[string]string `json:"fpxx"`
}

func Query(r render.Render, res http.ResponseWriter, req *http.Request) {

	data := make(map[string]interface{})
	data["isQuery"] = true

	r.HTML(200, "query", data)
}

func QueryWs(r render.Render, res http.ResponseWriter, req *http.Request) {

	conf := ini.DumpAll("conf/app.conf")

	req.ParseForm()

	key := req.FormValue("key")
	kpyf := req.FormValue("kpyf")

	url := conf["proxy:"+kpyf]

	rs, err := redis.Dial("tcp", url)
	if err != nil {
		log.Println(err)
	}
	defer rs.Close()

	v, err := redis.Values(rs.Do("HGETALL", key))
	if err != nil {
		panic(err)
	}

	m, err := redis.StringMap(v, err)

	output, _ := json.Marshal(&m)
	fmt.Fprintln(res, string(output))

}

type CYJG struct {
	Result string `json:"result"`
	Xfsbh  string `json:"xfsbh"`
	Xfmc   string `json:"xfmc"`
	Gfsbh  string `json:"gfsbh"`
	Gfmc   string `json:"gfmc"`
}

func QueryWs2(r render.Render, res http.ResponseWriter, req *http.Request) {

	conf := ini.DumpAll("conf/app.conf")

	req.ParseForm()

	key := req.FormValue("key")
	kpyf := req.FormValue("kpyf")

	url := conf["proxy:"+kpyf]

	rs, err := redis.Dial("tcp", url)
	if err != nil {
		log.Println(err)
	}
	defer rs.Close()

	v, err := redis.Values(rs.Do("HGETALL", key))
	if err != nil {
		panic(err)
	}

	m, err := redis.StringMap(v, err)

	xxx := CYJG{"1", m["06"], m["07"], m["02"], m["03"]}
	fmt.Println(xxx)

	output, _ := json.Marshal(&xxx)
	fmt.Println(string(output))
	fmt.Fprintln(res, string(output))

}
