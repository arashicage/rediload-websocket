package controllers

import (
	"fmt"
	"net/http"

	"common/goracle/connect"
	"github.com/martini-contrib/render"

	"common/ini"
	// "rediload-websocket/schedule"
)

func Tasks(r render.Render, res http.ResponseWriter, req *http.Request) {

	data := make(map[string]interface{})

	data["title"] = "查询待加载项"
	data["isSched"] = true

	recordes := retriveJobs()
	data["data"] = recordes

	r.HTML(200, "tasks", data)
}

func retriveJobs() [][]interface{} {

	conf := ini.DumpAll("conf/app.conf")

	uid := conf["common:uid"]

	sql := conf["common:sql"]

	cxSource, err := connect.GetRawConnection(uid)
	if err != nil {
		fmt.Println("get connection fail %q: %s", uid, err)
		return nil
	}
	defer cxSource.Close()

	cuSource := cxSource.NewCursor()
	defer cuSource.Close()

	err = cuSource.Execute(sql, nil, nil)
	if err != nil {

		fmt.Println("execute sql fail %q: %s", sql, err)
		return nil
	}

	records, err := cuSource.FetchAll()
	if err != nil {
		fmt.Println("获取结果集失败")
		return nil
	}
	return records
}
