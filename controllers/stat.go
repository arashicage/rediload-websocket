package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	// "github.com/Unknwon/goconfig"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/websocket"
	"github.com/martini-contrib/render"
	"gopkg.in/ini.v1"
)

type RedisStats map[string]map[string]string

type InfoMessage struct {
	Type  string `json:"type"`
	Msg   string `json:"msg"`
	Inf   string `json:"inf"`
	Chart string `json:"chart"`
}

func Stat(r render.Render, res http.ResponseWriter, req *http.Request) {

	data := make(map[string]interface{})

	data["title"] = "Redis 监控"
	data["isStat"] = true

	data["url"] = "127.0.0.1:6379"

	data["head"] = []string{ //history table head
		"time",
		"us",
		"sy",
		"cl",
		"bcl",
		"mem",
		"rss",
		"keys",
		"cmd/s",
		"exp/s",
		"evt/s",
		"hit%/s",
		"hit/s",
		"mis/s",
		"aofcs",
	}

	type info struct {
		Key, Val string
	}

	data["info"] = []info{ //redis information head
		info{"redis_version", "-"},
		info{"process_id", "-"},
		info{"uptime_in_secends", "-"},
		info{"uptime_in_days", "-"},
		info{"gcc_version", "-"},
		info{"role", "-"},
		info{"connected_slaves", "-"},
		info{"aof_enabled", "-"},
		info{"vm_enabled", "-"},
	}

	conf, err := ini.Load("conf/app.conf")
	if err != nil {
		fmt.Println("can not load file: conf/app.conf")
	}

	redis_urls := conf.Section("common").Key("redis_urls").String()

	data["redis_urls"] = strings.Split(redis_urls, ",")

	r.HTML(200, "stat", data)

}

func StatWs(r render.Render, res http.ResponseWriter, req *http.Request) {

	req.ParseForm()

	url := req.FormValue("url")

	interval := req.FormValue("interval")

	interval64, err := strconv.ParseUint(interval, 10, 64)
	if err != nil {
		log.Println(err)
	}

	// =======================================================================================

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

	us_last, sy_last := 0, 0
	us_curr, sy_curr := 0, 0
	us, sy := 0, 0

	// Message receive loop.
	for i := 0; ; i++ {

		//页面关闭还在运行

		//=======================================================

		info, err := QueryStats(url) // 要加异常处理, 连接不同的时候...
		if err != nil {
			break
		}

		stat := FiltrateStats(info, "used_cpu_user,used_cpu_sys,connected_clients,blocked_clients,used_memory,used_memory_rss,db0.keys,total_commands_processed,db0.expires,evicted_keys,keyspace_hit_ratio,keyspace_hits,keyspace_misses,aof_current_size,redis_version,process_id,uptime_in_seconds,uptime_in_days,gcc_version,role,connected_slaves,aof_enabled,vm_enabled")

		fmt.Println(stat, "\n")

		uptime_in_seconds, _ := strconv.Atoi(stat["uptime_in_seconds"])
		total_commands_processed, _ := strconv.Atoi(stat["total_commands_processed"])
		total_commands_processed_per_second := total_commands_processed / uptime_in_seconds
		expired_keys, _ := strconv.Atoi(stat["db0.expires"])
		evicted_keys, _ := strconv.Atoi(stat["evicted_keys"])
		expired_keys_per_second := float64(expired_keys) / float64(uptime_in_seconds)
		evicted_keys_per_second := float64(evicted_keys) / float64(uptime_in_seconds)

		keyspace_hits, _ := strconv.Atoi(stat["keyspace_hits"])
		keyspace_misses, _ := strconv.Atoi(stat["keyspace_misses"])
		keyspace_total := keyspace_hits + keyspace_misses
		keyspace_hit_ratio := float64(keyspace_hits) / float64(keyspace_total) * 100

		keyspace_hits_per_second := float64(keyspace_hits) / float64(uptime_in_seconds)
		keyspace_misses_per_second := float64(keyspace_misses) / float64(uptime_in_seconds)

		used_memory, _ := strconv.Atoi(stat["used_memory"])
		used_memory_MB := float64(used_memory) / 1024 / 1024

		used_memory_rss, _ := strconv.Atoi(stat["used_memory_rss"])
		used_memory_rss_MB := float64(used_memory_rss) / 1024 / 1024

		keys, _ := strconv.Atoi(stat["db0.keys"])
		keys_K := float64(keys) / 1000

		us_curr, _ = strconv.Atoi(stat["used_cpu_us"])
		sy_curr, _ = strconv.Atoi(stat["used_cpu_sys"])

		if i != 0 {
			us = us_curr - us_last
			sy = sy_curr - sy_last

		} else {
			us = 0
			sy = 0
		}

		us_last = us_curr
		sy_last = sy_curr

		message := InfoMessage{
			Type: "output",
			Msg: "<tr>" +
				"<td >" + time.Now().String()[11:19] + "</td>" +
				"<td style=color:orange>" + strconv.Itoa(us) + "</td>" +
				"<td style=color:orange>" + strconv.Itoa(sy) + "</td>" +
				"<td style=color:darkcyan>" + stat["connected_clients"] + "</td>" +
				"<td style=color:darkcyan>" + stat["blocked_clients"] + "</td>" +
				"<td style=color:green>" + strconv.FormatFloat(used_memory_MB, 'f', 1, 64) + "MB" + "</td>" +
				"<td style=color:green>" + strconv.FormatFloat(used_memory_rss_MB, 'f', 1, 64) + "MB" + "</td>" +
				"<td >" + strconv.FormatFloat(keys_K, 'f', 1, 64) + "K" + "</td>" +
				"<td style=color:blue>" + strconv.Itoa(total_commands_processed_per_second) + "</td>" +
				"<td style=color:red>" + strconv.FormatFloat(expired_keys_per_second, 'f', 4, 64) + "</td>" +
				"<td style=color:red>" + strconv.FormatFloat(evicted_keys_per_second, 'f', 6, 64) + "</td>" +
				"<td style=color:mediumorchid>" + strconv.FormatFloat(keyspace_hit_ratio, 'f', 4, 64) + "</td>" +
				"<td style=color:mediumorchid>" + strconv.FormatFloat(keyspace_hits_per_second, 'f', 4, 64) + "</td>" +
				"<td style=color:mediumorchid>" + strconv.FormatFloat(keyspace_misses_per_second, 'f', 4, 64) + "</td>" +
				"<td style=color:darkcyan>" + nvl(stat["aof_current_size"]) + "</td>" +
				"</tr>",
			Inf: url + "," +
				stat["redis_version"] + "," +
				stat["process_id"] + "," +
				stat["uptime_in_seconds"] + "," +
				stat["uptime_in_days"] + "," +
				stat["gcc_version"] + "," +
				stat["role"] + "," +
				stat["connected_slaves"] + "," +
				stat["aof_enabled"] + "," +
				nvl(stat["vm_enabled"]),
			Chart: time.Now().String()[:19] + "," +
				strconv.Itoa(total_commands_processed_per_second) + "," +
				stat["used_cpu_user"] + "," +
				stat["used_cpu_sys"] + "," +
				strconv.FormatFloat(used_memory_MB, 'f', 1, 64) + "," +
				strconv.FormatFloat(used_memory_rss_MB, 'f', 1, 64),
		}

		ws.WriteJSON(message)

		_, p, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("======", err)
			break //如果前台关闭页面,这里会抛出错误
		}

		fmt.Println(string(p), "\n")

		if string(p) == "alive" {

		} else if string(p) == "stop" {
			break
		}

		time.Sleep(time.Duration(interval64) * time.Second)

	}

}

func nvl(s string) string {
	ret := "N/A"
	if s != "" {
		ret = s
	}
	return ret
}

func QueryStats(host string) (RedisStats, error) {
	if c, err := redis.Dial("tcp", host); err == nil {
		defer c.Close()
		if stats, err := redis.String(c.Do("INFO")); err == nil {
			lines := strings.Split(stats, "\r\n")
			rs := RedisStats(make(map[string]map[string]string))
			var wg map[string]string
			var section string
			for _, l := range lines {
				if len(l) == 0 {
					wg = nil
					continue
				} else if strings.HasPrefix(l, "#") {
					k := strings.TrimSpace(strings.TrimPrefix(l, "#"))
					wg = make(map[string]string)
					rs[k] = wg
					section = k
				} else if wg != nil {
					if values := strings.Split(l, ":"); len(values) == 2 {
						if section == "Keyspace" {
							csv := strings.Split(values[1], ",")
							for _, query := range csv {
								if kvs := strings.Split(query, "="); len(kvs) == 2 {
									if _, err := strconv.ParseFloat(kvs[1], 64); err == nil {
										wg[values[0]+"."+kvs[0]] = kvs[1]
									}
								}
							}
						} else {
							wg[values[0]] = values[1]
							// if _, err := strconv.ParseFloat(values[1], 64); err == nil {
							// switch values[0] {
							// case "redis_git_sha1", "redis_git_dirty", "redis_build_id", "arch_bits",
							// 	"process_id", "tcp_port", "aof_enabled":
							// 	// don't report these metrics
							// 	continue
							// default:
							// wg[values[0]] = values[1]
							// }
							// }
						}
					}
				}
			}
			return rs, nil
		} else {
			log.Printf("Failed to query stats from redis @ %s. Error: %s", host, err.Error())
			return nil, err
		}
	} else {
		log.Printf("Failed to connect to redis @ %s. Error: %s", host, err.Error())
		return nil, err
	}
}

func FiltrateStats(s RedisStats, filter string) map[string]string {

	stat := make(map[string]string)

	// 当 filter 为 "" 时, 返回全部
	// 当 filter 为 "key1,key2,key3 ...", 返回 key1 key2 key3 对应的 value
	keys := strings.Split(filter, ",")

	for _, detail := range s {
		for k, v := range detail {
			if filter == "" {
				stat[k] = v
			} else {
				for _, key := range keys {
					if key == k {
						stat[k] = v
					}
				}
			}

		}
	}
	return stat
}
