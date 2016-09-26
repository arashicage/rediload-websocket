package controllers

import (
	"net/http"

	"github.com/martini-contrib/render"

	_ "bytes"
	"fmt"
	"github.com/RaymondChou/goqr/pkg"
	_ "image"
	_ "image/color"
	_ "image/png"
	"strings"
)

func Qr1(r render.Render, res http.ResponseWriter, req *http.Request) {

	data := make(map[string]interface{})

	data["title"] = "生成二维码"
	data["isQr1"] = true

	r.HTML(200, "qr1", data)
}

func QrApi(res http.ResponseWriter, req *http.Request) {

	// go run main.go -server=true
	// http://localhost:8889/?&data=%E5%BC%A0%E4%B8%89%E6%9D%8E%E5%9B%9B%E7%8E%8B%E4%BA%8C%E9%BA%BB
	var vaild bool

	req.ParseForm()

	l := req.FormValue("level")

	var qrl = qr.M

	switch l {
	case "L":
		qrl = qr.L
	case "M":
		qrl = qr.M
	case "Q":
		qrl = qr.Q
	case "H":
		qrl = qr.H
	}

	fmt.Println("================= redundant level", l)

	for k, v := range req.Form {

		if k == "data" {
			fmt.Println(k)
			fmt.Println(strings.Join(v, ""))

			data := strings.Join(v, "")

			// c, err := qr.Encode(data, qr.H)
			c, err := qr.Encode(data, qrl)
			if err != nil {
				fmt.Println(err)
			}
			pngdat := c.PNG()

			res.Header().Set("Content-Type", "image/png")
			res.Write(pngdat)

			defer func() {
				vaild = true
			}()
		}

	}

	if vaild == false {
		fmt.Fprintf(res, "Please input data using get method!")
	}
}
