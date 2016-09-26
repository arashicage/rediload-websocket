package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"common/goracle/connect"
	"common/ini"
	"github.com/martini-contrib/render"
)

func Qr2(r render.Render, res http.ResponseWriter, req *http.Request) {
	data := make(map[string]interface{})

	data["title"] = "qr2"

	conf := ini.DumpAll("conf/app.conf")

	uid := conf["common:uid"]

	sql := conf["qrcode:sql_01_01"]

	col := conf["qrcode:col_01_01"]

	qrs := conf["qrcode:qrs_01_01"]

	recordes := retriveQR2(uid, sql)
	data["data"] = recordes
	data["col"] = strings.Split(col, ",")

	data["qrs"] = qrs

	fmt.Println(data["col"])

	r.HTML(200, "qr2", data)
}

func retriveQR2(uid, sql string) [][]interface{} {

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
