package controllers

import (
	"fmt"
	"net/http"

	"github.com/martini-contrib/render"
	"gopkg.in/ini.v1"
)

func Conf(r render.Render, res http.ResponseWriter, req *http.Request) {

	data := make(map[string]interface{})

	conf, err := ini.Load("conf/app.conf")
	if err != nil {
		fmt.Println("can not found file: app.conf", err)
	}

	redis_urls := conf.Section("common").Key("redis_urls").String()

	mongo_urls := conf.Section("common").Key("mongo_urls").String()

	level_urls := conf.Section("common").Key("level_urls").String()

	cron := conf.Section("common").Key("cron").String()

	data["redis_urls"] = redis_urls
	data["mongo_urls"] = mongo_urls
	data["level_urls"] = level_urls
	data["cron"] = cron

	r.HTML(200, "conf", data)

}
