package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/garyburd/redigo/redis"
	"github.com/martini-contrib/render"
)

func Verify(r render.Render, res http.ResponseWriter, req *http.Request) {
	data := make(map[string]interface{})

	type xx struct {
		Name string
		Sex  string
	}

	x := xx{
		Name: "Jim",
		Sex:  "男fasdf sfasdf asdfsdfasd",
	}

	y := []int{1, 2, 3, 4, 5}

	data["x"] = x
	data["y"] = y
	data["title"] = "Verify"

	r.HTML(200, "v", data)
}

func Result1(r render.Render, res http.ResponseWriter, req *http.Request, log *log.Logger) {

	rs, err := redis.Dial("tcp", "192.168.5.102:6387")
	if err != nil {
		log.Println(err)
	}
	defer rs.Close()

	data := make(map[string]interface{})
	data["title"] = "查验结果"
	data["fpdm"] = req.FormValue("fpdm")
	data["fphm"] = req.FormValue("fphm")

	//	v, err := redis.Values(rs.Do("HGETALL", "01:110012414002573318")) //"01:"+data["fpdm"].(string)+data["fphm"].(string)))
	v, err := redis.Values(rs.Do("HGETALL", "01:"+data["fpdm"].(string)+data["fphm"].(string)))

	if err != nil {
		panic(err)
	}

	m, err := redis.StringMap(v, err)

	for k, v := range m {
		log.Println(k, v)
	}

	if m["5"] != "" && //一致
		req.FormValue("kprq") == m["5"] &&
		req.FormValue("je") == m["6"] &&
		req.FormValue("xfsbh") == m["1"] {
		data["je"] = m["6"]
		data["xfsbh"] = m["1"]
		data["xfmc"] = m["2"]
		data["kprq"] = m["5"]
		data["se"] = m["7"]
		data["gfsbh"] = m["3"]
		data["gfmc"] = m["4"]
		data["same"] = "一致"
	} else {
		data["same"] = "查无此票"
	}

	if m["8"] == "" {
		data["cnt"] = 1
	} else {
		tmp, _ := strconv.Atoi(m["8"])
		data["cnt"] = tmp + 1
	}

	log.Println(m["8"])

	rs.Do("HSET", "01:"+data["fpdm"].(string)+data["fphm"].(string), "8", data["cnt"])

	r.HTML(200, "result", data)

}

func Res(r render.Render, res http.ResponseWriter, req *http.Request, log *log.Logger) {

	rs, err := redis.Dial("tcp", "172.30.11.230:6379")
	if err != nil {
		log.Println(err)
	}
	defer rs.Close()

	data := make(map[string]interface{})
	data["title"] = "查验结果"
	data["fpdm"] = req.FormValue("fpdm")
	data["fphm"] = req.FormValue("fphm")

	//	v, err := redis.Values(rs.Do("HGETALL", "01:110012414002573318")) //"01:"+data["fpdm"].(string)+data["fphm"].(string)))
	v, err := redis.Values(rs.Do("HGETALL", "01:"+data["fpdm"].(string)+data["fphm"].(string)))

	if err != nil {
		panic(err)
	}

	m, err := redis.StringMap(v, err)

	for k, v := range m {
		log.Println(k, v)
	}

	data["je"] = m["6"]
	data["xfsbh"] = m["1"]
	data["kprq"] = m["5"]
	data["se"] = m["7"]
	data["gfsbh"] = m["3"]

	// 1 xfsbh
	// 2 xfmc
	// 3 gfsbh
	// 4 gfmc
	// 5 kprq
	// 6 je
	// 7 se

	r.HTML(200, "res", data)
}
