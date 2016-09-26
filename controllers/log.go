package controllers

import (
	// "fmt"
	"net/http"
	// "runtime"
	// "strconv"
	// "strings"
	// "time"

	// "code.google.com/p/mahonia"
	// "github.com/Unknwon/goconfig"
	// "github.com/garyburd/redigo/redis"
	// "github.com/gorilla/websocket"
	"github.com/martini-contrib/render"

	"fmt"

	"common/goracle/connect"

	"common/ini"
)

func Log(r render.Render, res http.ResponseWriter, req *http.Request) {
	data := make(map[string]interface{})

	data["title"] = "调度日志"
	data["isSched"] = true

	recordes := retriveLogs()
	data["data"] = recordes

	r.HTML(200, "logs", data)
}

func retriveLogs() [][]interface{} {

	conf := ini.DumpAll("conf/app.conf")

	uid := conf["common:uid"]

	sql := conf["common:sql_log"]

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
