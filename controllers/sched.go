package controllers

import (
	"fmt"
	"net/http"

	"github.com/Unknwon/goconfig"
	"github.com/martini-contrib/render"
	"gopkg.in/ini.v1"
	"rediload-websocket/modules/rediload"
	"rediload-websocket/modules/schedule"

)

var f = "conf/app.conf"

func Sched(r render.Render, res http.ResponseWriter, req *http.Request) {
	data := make(map[string]interface{})

	data["title"] = "调度设置"
	data["isSched"] = true
	data["header"] = "下次调度时间"

	conf, err := ini.Load(f)
	if err != nil {
		fmt.Println("can not load file:", f)
	}

	cron := conf.Section("common").Key("cron").String()

	data["cron"] = cron

	em := schedule.C.EntrySnapshotMap()
	task := em["MainCron"]
	data["maincron"] = task.Next.String()[:19]
	fmt.Println(task.Next.String())

	r.HTML(200, "sched", data)
}

func SaveCron(r render.Render, res http.ResponseWriter, req *http.Request) {

	req.ParseForm()

	cron := req.FormValue("cron")

	conf, err := goconfig.LoadConfigFile(f)
	if err != nil {
		fmt.Println("can not load file:", f)
	}

	conf.SetValue("common", "cron", cron)

	goconfig.SaveConfigFile(conf, f)

	schedule.C.RemoveByName("MainCron")
	schedule.C.AddFunc(cron, rediload.Cli, "MainCron")
	schedule.C.Start()

	for k, j := range schedule.C.EntrySnapshotMap() {
		fmt.Println(k, "----------", j.Next.String()[:19])
	}

	r.Redirect("/sched")

}
