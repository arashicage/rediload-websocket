package main

import (
	"rediload-websocket/controllers"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"

	"rediload-websocket/modules/schedule" //init 创建自动任务

	"github.com/codegangsta/martini-contrib/web"

	"fmt"
)

const VERSION = "0.0.01.0525"
const GO_VERSION = "go version go1.5.4 darwin/amd64"

var (
	m *martini.ClassicMartini
)

func printCompileEnv(){
	fmt.Println("compiled with",GO_VERSION)
	fmt.Println("current version",VERSION)
}

func setTemplates(t string) martini.Handler {
	if t == "" { //不使用layout
		return render.Renderer(render.Options{
			Directory:  "templates",
			Extensions: []string{".html", ".tmpl"},
		})
	} else { //使用指定layout
		return render.Renderer(render.Options{
			Directory:  "templates",
			Extensions: []string{".html", ".tmpl"},
			Layout:     t,
		})
	}
}

func main() {

	printCompileEnv()

	// 初始化配置信息
	// init

	// tmpl := "base/layout"
	tmpl := "base/starter"

	m = martini.Classic()

	m.Map(schedule.C)

	m.Use(martini.Static("www")) //静态页面

	// m.Use(setTemplates(tmpl))

	m.Get("/", setTemplates(tmpl), controllers.Load)

	m.Get("/load", setTemplates(tmpl), controllers.Load)

	m.Get("/sched", setTemplates(tmpl), controllers.Sched)

	m.Get("/scron", setTemplates(tmpl), controllers.SaveCron)

	m.Get("/tasks", setTemplates(tmpl), controllers.Tasks)

	m.Get("/log", setTemplates(tmpl), controllers.Log)

	m.Get("/query", setTemplates(tmpl), controllers.Query)

	m.Get("/query/ws", setTemplates(tmpl), controllers.QueryWs)

	m.Get("/query/ws2", setTemplates(tmpl), controllers.QueryWs2)

	m.Get("/options", setTemplates(tmpl), controllers.Options)

	m.Get("/load/ws", setTemplates(tmpl), controllers.LoadWs)

	// ====== redis-stat
	m.Get("/stat", setTemplates(tmpl), controllers.Stat)

	m.Get("/stat/ws", setTemplates(tmpl), controllers.StatWs)

	// ====== scan
	m.Get("/scan", setTemplates(tmpl), controllers.Scan)

	m.Get("/scan/ws", setTemplates(""), controllers.ScanWs)

	m.Get("/locate", setTemplates(tmpl), controllers.Locate)

	m.Get("/locate/ws", setTemplates(tmpl), controllers.LocateWsFromLevelDB)

	// ====== conf
	m.Get("/conf", setTemplates(tmpl), controllers.Conf)

	m.Get("/save", setTemplates(tmpl), controllers.SaveConf)

	// ====== help about etc
	m.Get("/help", setTemplates(tmpl), controllers.Help)

	m.Get("/about", setTemplates(tmpl), controllers.About)

	// ====== QR begin ======
	m.Get("/qr", setTemplates(""), controllers.QrApi)

	m.Get("/qr1", setTemplates(tmpl), controllers.Qr1)

	m.Get("/qr2", setTemplates(tmpl), controllers.Qr2)
	// ====== QR end ======

	// ======================================
	m.Use(web.ContextWithCookieSecret(""))

	// 这里实现注册检查和注册处理的工作
	m.Use(func() {
		fmt.Println("hello world=======")
	})

	// ======================================listen on port 8080
	m.RunOnAddr(":8080")
}
