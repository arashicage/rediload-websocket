package controllers

import (
	"net/http"

	"github.com/martini-contrib/render"
	"github.com/satori/go.uuid"
	"log"
	"strings"
)

func Index(r render.Render) {

	r.Redirect("/index.html")
}

func Load(r render.Render, res http.ResponseWriter, req *http.Request) {
	data := make(map[string]interface{})

	u1 := uuid.NewV4()

	data["title"] = "Redis 加载"
	data["log"] = u1.String() + ".log"
	data["isLoad"] = true

	r.HTML(200, "load", data)
}

func LoadWs(r render.Render, res http.ResponseWriter, req *http.Request) {
	data := make(map[string]interface{})

	req.ParseForm()

	data["what"] = req.FormValue("what")
	data["url"] = req.FormValue("url")
	data["core"] = req.FormValue("core")
	data["worker"] = req.FormValue("worker")
	data["uid"] = req.FormValue("uid")
	sql := strings.Replace(req.FormValue("sql"), "_@", "\r\n", -1)
	sql = strings.Replace(sql, "_$", "+", -1)
	data["sql"] = sql
	data["batch"] = req.FormValue("batch")
	data["fields"] = req.FormValue("fields")
	data["codepage"] = req.FormValue("codepage")
	data["log"] = req.FormValue("log")

	log.Println(data["sql"])

	data["title"] = "Redis 加载"

	ws(res, req, data)

}
