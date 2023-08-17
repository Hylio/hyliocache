# hyliocache
用go实现类groupchache。

# Highlight
1. 使用RPC框架取代http框架进行通信

# Usage
```
package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hylio/hyliocache"
	"log"
	"net/http"
)

var db = map[string]string{
	"zhanghao":    "hylio",
	"wangrui":     "civet",
	"zhouruqiang": "dio",
}

func startCacheServer(addr string, addrs []string, g *hyliocache.Group) {
	peers := hyliocache.NewHTTPPool(addr)
	peers.Set(addrs...)
	g.RegisterPeers(peers)
	log.Println("hylioCache is running at", addr)
	r := gin.Default()
	r.GET("_hyliocache/:gorupName/:key", peers.Serve)
	log.Fatal(r.Run(addr[7:]))
}

func startApiServer(apiAddr string, g *hyliocache.Group) {
	r := gin.Default()
	r.GET("/api", func(c *gin.Context) {
		key := c.DefaultQuery("key", "")
		view, err := g.Get(key)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.Header("Content-Type", "application/octet-stream")
		c.Data(http.StatusOK, "application/octet-stream", view.ByteSlice())
	})
	log.Println("frontend server is running at", apiAddr)
	log.Fatal(r.Run(apiAddr[7:]))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "hyliocache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	c := hyliocache.NewGroup("aka", 2<<10, hyliocache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[DB] searching key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	if api {
		go startApiServer(apiAddr, c)
	}
	startCacheServer(addrMap[port], addrs, c)
}

```
