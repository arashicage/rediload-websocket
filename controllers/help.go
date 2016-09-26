package controllers

import (
	"net/http"

	"github.com/martini-contrib/render"
)

func Help(r render.Render, res http.ResponseWriter, req *http.Request) {
	data := make(map[string]interface{})

	data["title"] = "参数说明"
	data["isHelp"] = true

	r.HTML(200, "help", data)
}
