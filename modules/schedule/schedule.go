package schedule

import (
	"common/cron"
	"common/ini"
	"log"
	"rediload-websocket/modules/rediload"
)

var (
	C *cron.Cron
)

func init() {

	m := ini.DumpAll("conf/app.conf")

	c := m["common:cron"]

	log.Println("自动任务 corn 表达式:", c)

	C = cron.New()

	C.AddFunc(c, rediload.Cli, "MainCron")

	C.Start()
}
