package hyliocache

//import (
//	"fmt"
//	"github.com/golang/protobuf/proto"
//	pb "github.com/hylio/hyliocache/hyliocachepb"
//	"io"
//	"net/http"
//	"net/url"
//)

import (
	"context"
	"fmt"
	pb "github.com/hylio/hyliocache/hyliocachepb"
	"github.com/hylio/hyliocache/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

// client 实现了访问其他远程节点并获取缓存的能力

type Client struct {
	addr string // 定义将要访问的服务的地址 ip:port
}

func (c *Client) Get(in *pb.Request) ([]byte, error) {
	// 创建一个etcd client
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	// 发现服务
	conn, err := registry.EtcdDial(cli, "_hyliocache" + "/" + c.addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	grpcClient := pb.NewGroupCacheClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	group, key := in.GetGroup(), in.GetKey()
	resp, err := grpcClient.Get(ctx, &pb.Request{
		Group: group,
		Key:   key,
	})

	if err != nil {
		return nil, fmt.Errorf("can not get %s/%s from peer %s", group, key, c.addr)
	}
	bytes := resp.GetValue()

	return bytes, nil
}

func NewClient(addr string) *Client {
	return &Client{addr: addr}
}

// 测试Client是否实现了PeerGetter接口
var _ PeerGetter = (*Client)(nil)
