// 代理模式下,不支持事务 MULTI 不支持 SCAN

package controllers

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/mahonia"
	"common/goracle"
	"common/goracle/connect"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/websocket"
	"gopkg.in/goracle.v1/oracle"
	"gopkg.in/ini.v1"
	// "rediload-websocket/common/goracle/connect"
)

type Args struct {
	hkey string
	hval string
}

type Para struct {
	url    string
	core   int
	worker int
	batch  int
	uid    string
	sql    string
	fields string
	gbk    string
	opt    string
}

type Result struct {
	id   int
	line int
}

type Message struct {
	Type string `json:"type"`
	Msg  string `json:"msg"`
}

type Job struct {
	id      int                          //批次
	records [][]interface{}              //oracle 结果集
	columns []goracle.Column             // 列字段信息
	desc    []oracle.VariableDescription // 字段列描述
	N       int                          //字段数量
	fields  []string                     //目标fields
	results chan<- Result                //本批次处理结果
}

func addJobs(jobs chan<- Job, para *Para, timeBegin time.Time, results chan<- Result, ws *websocket.Conn) { /* ([]Column, []oracle.VariableDescription, error) */

	cxSource, err := connect.GetRawConnection(para.uid)
	defer cxSource.Close()
	if err != nil {

		message := Message{
			Type: "error",
			Msg:  fmt.Sprintf("connect to database fail :", err),
		}
		ws.WriteJSON(message)

		fmt.Println("connect to database fail :", err)
		return
	}

	cuSource := cxSource.NewCursor()
	defer cuSource.Close()

	err = cuSource.Execute(para.sql, nil, nil)
	if err != nil {

		message := Message{
			Type: "error",
			Msg:  fmt.Sprintf("execute sql fail %q: %s", para.sql, err),
		}
		ws.WriteJSON(message)

		fmt.Println("execute sql fail %q: %s", para.sql, err)
		return
	}

	columns, err := goracle.GetColumns(cuSource)
	if err != nil {

		message := Message{
			Type: "error",
			Msg:  fmt.Sprintf("retrieve columns fail :%s", err),
		}
		ws.WriteJSON(message)

		fmt.Println("retrieve columns fail :%s", err)
		return
	}

	fieldnames := strings.Split(para.fields, ",")

	desc, err := cuSource.GetDescription()

	N := cap(desc) //字段数量
	if N != len(fieldnames) {

		message := Message{
			Type: "error",
			Msg:  fmt.Sprintf("the count of fields in sql and fields not the same "),
		}
		ws.WriteJSON(message)

		fmt.Println("the count of fields in sql and fields not the same ")
		return
	}

	id := 1
	for records, err := cuSource.FetchMany(para.batch); err == nil && len(records) > 0; records, err = cuSource.FetchMany(para.batch) {

		message := Message{
			Type: "output",
			Msg:  fmt.Sprintf(">>> retrive %5d group of data \n--- %s   %-18s   %d", id, time.Now().String()[0:27], time.Now().Sub(timeBegin), para.batch*(id-1)+len(records)),
		}
		ws.WriteJSON(message)

		fmt.Printf(">>> retrive %5d group of data --- %s   %-18s   %d\n", id, time.Now().String()[0:27], time.Now().Sub(timeBegin), para.batch*(id-1)+len(records))
		jobs <- Job{id, records, columns, desc, N, fieldnames, results}
		id = id + 1
	}
	close(jobs)
}

func doJobs(done chan<- struct{}, jobs <-chan Job, para *Para, timeBegin time.Time, ws *websocket.Conn) {
	for job := range jobs {
		// fmt.Println(job.id, len(job.records))

		// strings.Contains(para.field, "expire")
		if para.opt == "01" {
			pipeHashFPXX(job, timeBegin, para, ws)
		}

		// if para.fields == "key,field,value"
		if para.opt == "02" {
			pipeHashHWXX(job, timeBegin, para, ws)
		}

		job.results <- Result{job.id, len(job.records)}
	}
	done <- struct{}{}
}

func ws(res http.ResponseWriter, req *http.Request, m map[string]interface{}) {

	// Upgrade from http request to WebSocket.
	ws, err := websocket.Upgrade(res, req, nil, 1024, 1024)
	defer ws.Close()
	if _, ok := err.(websocket.HandshakeError); ok {
		fmt.Println(res, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		fmt.Println("Cannot setup WebSocket connection:", err)
		return
	}

	dec := func(x string) string {
		var y string
		if x == "01" {
			y = "01 发票信息"
		} else if x == "02" {
			y = "02 货物信息"
		}
		return y
	}

	message := Message{
		Type: "output",
		Msg: "--- parameters info -----------------------------------------------" + "\n" +
			"- what        : " + dec(m["what"].(string)) + "\n" +
			"- url         : " + m["url"].(string) + "\n" +
			"- core        : " + m["core"].(string) + "\n" +
			"- worker      : " + m["worker"].(string) + "\n" +
			"- uid         : " + m["uid"].(string) + "\n" +
			"- sql         : " + m["sql"].(string) + "\n" +
			"- fields      : " + m["fields"].(string) + "\n" +
			"- batch       : " + m["batch"].(string) + "\n" +
			"- codepage    : " + m["codepage"].(string) + "\n" +
			"- log         : " + m["log"].(string) + "\n" +
			"--- parameters info -----------------------------------------------",
	}
	ws.WriteJSON(message)

	timeBegin := time.Now()
	fmt.Println("=================================================================")

	// para, err := parseConf("redigo.conf")
	// if err != nil {
	// 	fmt.Println("retrieve parameters fail,please check your configuration file: redigo.conf")
	// 	return
	// }

	core_, _ := strconv.Atoi(m["core"].(string))
	worker_, _ := strconv.Atoi(m["worker"].(string))
	batch_, _ := strconv.Atoi(m["batch"].(string))

	para := &Para{
		url:    m["url"].(string),
		core:   core_,
		worker: worker_,
		batch:  batch_,
		uid:    m["uid"].(string),
		sql:    m["sql"].(string),
		fields: m["fields"].(string),
		gbk:    m["codepage"].(string),
		opt:    m["what"].(string),
	}

	min := minimum(para.core, runtime.NumCPU())
	runtime.GOMAXPROCS(min)

	message = Message{
		Type: "output",
		Msg:  fmt.Sprintf("%10s  :  %v", "core used", min),
	}
	ws.WriteJSON(message)

	message = Message{
		Type: "output",
		Msg:  fmt.Sprintf("================================================================="),
	}
	ws.WriteJSON(message)

	fmt.Printf("%10s  :  %v\n", "core used", min)
	fmt.Println("=================================================================")

	workers := para.worker
	jobs := make(chan Job, workers)       //遍历数据库，切分成 N 个 Job
	results := make(chan Result, workers) //每个小任务处理的结果
	done := make(chan struct{}, workers)  //每个小任务是否完成

	fmt.Printf("begin time                      --- %s\n", timeBegin.String()[0:27])
	go addJobs(jobs, para, timeBegin, results, ws)
	for i := 0; i < workers; i++ {
		go doJobs(done, jobs, para, timeBegin, ws) // Each executes in its own goroutine
	}
	go awaitCompletion(done, workers, results)
	processResults(results) // Blocks until the work is done

	message = Message{
		Type: "completed",
		Msg:  "--- completed -----------------------------------------------",
	}
	ws.WriteJSON(message)
}

func awaitCompletion(done <-chan struct{}, workers int, results chan Result) {
	for i := 0; i < workers; i++ {
		<-done
	}
	close(results)
}

func processResults(results <-chan Result) {
	for _ = range results {
	}
}

func pipeHashFPXX(job Job, timeBegin time.Time, para *Para, ws *websocket.Conn) {

	enc := mahonia.NewEncoder("GBK")

	rs, err := redis.Dial("tcp", para.url)

	if err != nil {

		message := Message{
			Type: "error",
			Msg:  fmt.Sprintf("ERROR routeine %5d create connection to redis:%s", job.id, err),
		}
		ws.WriteJSON(message)

		fmt.Printf("ERROR routeine %5d create connection to redis:%s\n", job.id, err)
	}

	for _, row := range job.records { //每行
		argx := make([]interface{}, job.N*2, job.N*2)
		idx := 0
		for i, col := range row { //每列
			r := ""
			if col == nil {
				r = ""
			} else {
				r = strings.TrimSuffix(strings.TrimPrefix(job.columns[i].String(col), `"`), `"`)
				if para.gbk == "gbk" {
					r = enc.ConvertString(r)
				}

			}
			argx[idx] = job.fields[i] //fieldnames
			idx++
			argx[idx] = r //field value
			idx++

		}
		//    1                            2n-2   2n-1
		//key xx field1 xx1 field2 xx2 ... expire nnn
		rs.Send("HMSET", argx[1:job.N*2-2]...)
		rs.Send("EXPIRE", argx[1], argx[job.N*2-1])
	}

	err = rs.Flush()
	for i := 0; i < len(job.records)*2; i++ {
		//		v, err := rs.Receive()
		_, err := rs.Receive()
		if err != nil {

			message := Message{
				Type: "error",
				Msg:  fmt.Sprintf("ERROR routeine %5d receive reply with err:%s", job.id, err),
			}
			ws.WriteJSON(message)

			fmt.Printf("ERROR routeine %5d receive reply with err:%s\n", job.id, err)

			break
		}
		//		fmt.Println(v)
	}

	if err != nil {

		message := Message{
			Type: "error",
			Msg:  fmt.Sprintf("ERROR routeine %5d loading data to redis:%s", job.id, err),
		}
		ws.WriteJSON(message)

		fmt.Printf("ERROR routeine %5d loading data to redis:%s\n", job.id, err)
	} else {

		message := Message{
			Type: "output",
			Msg:  fmt.Sprintf("<<< loading %5d group of data \n--- %s   %-18s   %-12d    routine:%5d", job.id, time.Now().String()[0:27], time.Now().Sub(timeBegin), len(job.records)+(job.id-1)*para.batch, job.id),
		}
		ws.WriteJSON(message)

		fmt.Printf("<<< loading %5d group of data --- %s   %-18s   %-12d    routine:%5d\n", job.id, time.Now().String()[0:27], time.Now().Sub(timeBegin), len(job.records)+(job.id-1)*para.batch, job.id)
	}

	rs.Close()
}

func pipeHashHWXX(job Job, timeBegin time.Time, para *Para, ws *websocket.Conn) {

	enc := mahonia.NewEncoder("GBK")

	rs, err := redis.Dial("tcp", para.url)

	if err != nil {

		message := Message{
			Type: "error",
			Msg:  fmt.Sprintf("ERROR routeine %5d create connection to redis:%s", job.id, err),
		}
		ws.WriteJSON(message)

		fmt.Printf("ERROR routeine %5d create connection to redis:%s\n", job.id, err)
	}

	for _, row := range job.records { //每行
		argx := make([]interface{}, job.N, job.N)
		idx := 0
		for i, col := range row { //每列
			r := ""
			if col == nil {
				r = ""
			} else {
				r = strings.TrimSuffix(strings.TrimPrefix(job.columns[i].String(col), `"`), `"`)
				if para.gbk == "gbk" {
					r = enc.ConvertString(r)
				}
			}
			argx[idx] = r //field value
			idx++

		}

		//key val
		rs.Send("HMSET", argx...)
		// rs.Send("EXPIRE", argx[1], argx[job.N*2-1])
	}

	err = rs.Flush()
	for i := 0; i < len(job.records); i++ {
		_, err := rs.Receive()
		if err != nil {

			message := Message{
				Type: "error",
				Msg:  fmt.Sprintf("ERROR routeine %5d receive reply with err:%s", job.id, err),
			}
			ws.WriteJSON(message)

			fmt.Printf("ERROR routeine %5d receive reply with err:%s\n", job.id, err)
		}
	}

	if err != nil {

		message := Message{
			Type: "error",
			Msg:  fmt.Sprintf("ERROR routeine %5d loading data to redis:%s", job.id, err),
		}
		ws.WriteJSON(message)

		fmt.Printf("ERROR routeine %5d loading data to redis:%s\n", job.id, err)
	} else {

		message := Message{
			Type: "output",
			Msg:  fmt.Sprintf("<<< loading %5d group of data \n--- %s   %-18s   %-12d    routine:%5d", job.id, time.Now().String()[0:27], time.Now().Sub(timeBegin), len(job.records)+(job.id-1)*para.batch, job.id),
		}
		ws.WriteJSON(message)

		fmt.Printf("<<< loading %5d group of data --- %s   %-18s   %-12d    routine:%5d\n", job.id, time.Now().String()[0:27], time.Now().Sub(timeBegin), len(job.records)+(job.id-1)*para.batch, job.id)
	}

	rs.Close()
}

func minimum(x int, others ...int) int {
	for _, y := range others {
		if y < x {
			x = y
		}
	}
	return x
}

func parseConf(fname string) (*Para, error) {

	para := &Para{}
	conf, err := ini.Load(fname)
	if err != nil {
		fmt.Println("can not found file: ", fname)
		return nil, err
	}

	url := conf.Section("DEFAULT").Key("url").String()
	core := conf.Section("DEFAULT").Key("core").MustInt()
	worker := conf.Section("DEFAULT").Key("worker").MustInt()
	batch := conf.Section("DEFAULT").Key("batch").MustInt()
	uid := conf.Section("DEFAULT").Key("uid").String()
	sql := conf.Section("DEFAULT").Key("sql").String()
	fields := conf.Section("DEFAULT").Key("fields").String()

	para.url = url
	para.core = core
	para.worker = worker
	para.batch = batch
	para.uid = uid
	para.sql = sql
	para.fields = strings.Replace(fields, " ", "", -1)

	fmt.Printf("parameters  :\n")
	fmt.Printf("%10s  :  %v\n", "url", para.url)
	fmt.Printf("%10s  :  %v\n", "core", para.core)
	fmt.Printf("%10s  :  %v\n", "worker", para.worker)
	fmt.Printf("%10s  :  %v\n", "batch", para.batch)
	fmt.Printf("%10s  :  %v\n", "uid", para.uid)
	fmt.Printf("%10s  :  %v\n", "sql", para.sql)
	fmt.Printf("%10s  :  %v\n", "fields", para.fields)

	return para, nil

}
