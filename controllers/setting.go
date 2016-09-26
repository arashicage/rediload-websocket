package controllers

import (
	"fmt"
	"net/http"

	"github.com/martini-contrib/render"
	"gopkg.in/ini.v1"
)

const CRON_DEFAULT = "* * 0/12 * *"

func SaveConf(r render.Render, res http.ResponseWriter, req *http.Request) {

	req.ParseForm()

	f := "conf/app.conf"

	redis_urls := req.FormValue("redisURLS")
	mongo_urls := req.FormValue("mongoURL")
	level_urls := req.FormValue("levelURLS")
	cron := req.FormValue("cron")

	cfg, err := ini.Load(f)
	if err != nil {
		fmt.Println("can not found file: app.conf")
	}

	if redis_urls != "" {
		cfg.Section("common").Key("redis_urls").SetValue(redis_urls)
	}

	if mongo_urls != "" {
		cfg.Section("common").Key("mongo_urls").SetValue(mongo_urls)
	}

	if level_urls != "" {
		cfg.Section("common").Key("level_urls").SetValue(level_urls)
	}

	// 如果传入了cron 参数
	if cron != "" {
		cfg.Section("common").Key("cron").SetValue(cron)
	}

	err = cfg.SaveTo(f)
	fmt.Println(err)
	if err != nil {
		fmt.Fprintln(res, `{"status":"fail"}`)
	} else {
		fmt.Fprintln(res, `{"status":"succ"}`)
	}

}
