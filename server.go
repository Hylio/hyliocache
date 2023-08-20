package hyliocache

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"github.com/hylio/hyliocache/consistenthash"
	pb "github.com/hylio/hyliocache/hyliocachepb"
	"github.com/hylio/hyliocache/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Server 模块提供了cache之间的通信能力
// peer节点之间可以通过server来获取其他节点的缓存

const (
	defaultAddr        = "127.0.0.1:4396"
	defaultBaseService = "/_hyliocache/"
	defaultReplicas    = 50
)

var (
	defaultEtcdConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	}
)

// Server 实现了服务端功能
type Server struct {
	pb.UnimplementedGroupCacheServer
	addr        string // 服务地址 like "http://localhost:8080"
	baseService string // 服务名称
	mu          sync.Mutex
	peers       *consistenthash.Map // 一致性哈希 选择节点
	clients     map[string]*Client  // 每个节点对应的client
	stopSignal  chan error          // 通知etcd 服务停止
	status      bool
}

func NewServer(addr string) *Server {
	if addr == "" {
		addr = defaultAddr
	}
	return &Server{
		addr:        addr,
		baseService: defaultBaseService,
	}
}

func (p *Server) Get(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	group, key := in.GetKey(), in.GetKey()
	resp := &pb.Response{}

	log.Printf("[hyliocache_svr %s] Receive RPC request - (%s)/(%s)", p.addr, group, key)
	if key == "" {
		return resp, fmt.Errorf("key is required")
	}
	g := GetGroup(group)
	if g == nil {
		return resp, fmt.Errorf("group not found")
	}
	view, err := g.Get(key)
	if err != nil {
		return resp, err
	}
	resp.Value = view.ByteSlice()
	return resp, nil
}

// Start 启动服务
func (p *Server) Start() error {
	p.mu.Lock()
	if p.status == true {
		p.mu.Unlock()
		return fmt.Errorf("server already start")
	}
	// 设置服务状态 添加报错通道
	p.status = true
	p.stopSignal = make(chan error)

	// 初始化tcp socket
	port := strings.Split(p.addr, ":")[1]
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// 注册rpc服务到grpc
	grpcServer := grpc.NewServer()
	pb.RegisterGroupCacheServer(grpcServer, p)

	// 注册到etcd
	go func() {
		err := registry.Registry("hyliocache", p.addr, p.stopSignal)
		if err != nil {
			log.Fatalf(err.Error())
		}
		close(p.stopSignal)
		err = lis.Close()
		if err != nil {
			log.Fatalf(err.Error())
		}
		log.Printf("[%s] Revoke service and close tcp socket", p.addr)
	}()

	p.mu.Unlock()

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}

func (p *Server) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.addr, fmt.Sprintf(format, v...))
}

func (p *Server) Serve(c *gin.Context) {
	// 限制访问路径
	if !strings.HasPrefix(c.Request.URL.Path, p.baseService) {
		panic("Server serving unexpected path: " + c.Request.URL.Path)
	}
	p.Log("%s %s", c.Request.Method, c.Request.URL.Path)
	// /<basepath>/<groupname>/<key>
	parts := strings.SplitN(c.Request.URL.Path[len(p.baseService):], "/", 2)
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
func (p *Server) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.clients = make(map[string]*Client, len(peers))
	for _, peer := range peers {
		if !CheckAddr(peer) {
			panic(fmt.Sprintf("[peer %s] is invalid!", peer))
		}
		p.clients[peer] = NewClient(peer + p.baseService)
	}
}

// PickPeer 根据一致性哈希找到key应该存放的节点 返回false说明应该从本地获取
func (p *Server) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.GetPeer(key); peer != "" && peer != p.addr {
		p.Log("Pick remote peer %s", peer)
		return p.clients[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*Server)(nil)
