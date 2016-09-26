package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"common/goracle"

	"github.com/martini-contrib/render"

	"gopkg.in/ini.v1"
)

// 存放加载数据的配置信息
type options struct {
	SQL    string `json:"sql"`
	Fields string `json:"fields"`
}

// 获取用于加载数据用的配置信息 sql 语句和 fields
func Options(r render.Render, res http.ResponseWriter, req *http.Request) {

	req.ParseForm()

	cfg, err := ini.Load("conf/app.conf")
	if err != nil {
		fmt.Println("can not load file: conf/app.conf")
	}

	uid := cfg.Section("common").Key("uid").String()
	sql := cfg.Section("options").Key("sql").String()
	col := cfg.Section("options").Key("col").String()
	key := cfg.Section("options").Key("key").String()
	del := cfg.Section("options").Key("del").String()

	m, _ := goracle.DumpTable(uid, sql, col, key, del)

	okey := req.FormValue("fplx") + del + req.FormValue("sjlx")

	op := options{m[okey]["sql_text"], m[okey]["cols"]}

	output, _ := json.Marshal(&op)
	fmt.Fprintln(res, string(output))

}
