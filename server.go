package hyliocache

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"github.com/hylio/hyliocache/consistenthash"
	pb "github.com/hylio/hyliocache/hyliocachepb"
	"log"
	"net/http"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_hyliocache/"
	defaultReplicas = 50
)

// HTTPPool 实现了服务端功能
type HTTPPool struct {
	self        string // 服务地址 like "http://localhost:8080"
	basePath    string // 服务路径
	mu          sync.Mutex
	peers       *consistenthash.Map    // 一致性哈希 选择节点
	httpGetters map[string]*httpGetter // 每个节点对应的httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) Serve(c *gin.Context) {
	// 限制访问路径
	if !strings.HasPrefix(c.Request.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + c.Request.URL.Path)
	}
	p.Log("%s %s", c.Request.Method, c.Request.URL.Path)
	// /<basepath>/<groupname>/<key>
	parts := strings.SplitN(c.Request.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	groupName, key := parts[0], parts[1]
	group := GetGroup(groupName)

	if group == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "no such group: " + groupName})
		return
	}

	view, err := group.Get(key)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	body, err := proto.Marshal(&pb.Response{
		Value: view.ByteSlice(),
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Type", "application/octet-stream")
	c.Writer.Write(body)
}

// Set 将各个远端地址配置到Server里
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)
