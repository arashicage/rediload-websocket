package controllers

import (
	"net/http"

	"github.com/martini-contrib/render"
)

func About(r render.Render, res http.ResponseWriter, req *http.Request) {
	data := make(map[string]interface{})

	data["title"] = "About"
	data["isAbout"] = true

	r.HTML(200, "about", data)
}
