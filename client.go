package hyliocache

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	pb "github.com/hylio/hyliocache/hyliocachepb"
	"github.com/hylio/hyliocache/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

// client 实现了访问其他远程节点并获取缓存的能力

type Client struct {
	addr string // 定义将要访问的服务的地址 ip:port
}

func (h *Client) Get(in *pb.Request, out *pb.Response) error {
	// 创建一个etcd client
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		return err
	}
	defer cli.Close()

	// 发现服务
	conn, err := registry.EtcdDial(cli, h.addr)
	if err != nil {
		return err
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
		return fmt.Errorf("can not get %s/%s from peer %s", group, key, h.addr)
	}
	bytes := resp.GetValue()

	//u := fmt.Sprintf("%v%v/%v", h.addr, url.QueryEscape(in.GetGroup()), url.QueryEscape(in.GetKey()))
	//res, err := http.Get(u)
	//if err != nil {
	//	return err
	//}
	//defer res.Body.Close()
	//
	//if res.StatusCode != http.StatusOK {
	//	return fmt.Errorf("Server returned： %v", res.Status)
	//}
	//
	//bytes, err := io.ReadAll(res.Body)
	//if err != nil {
	//	return fmt.Errorf("reading response body: %v", err)
	//}
	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}

func NewClient(addr string) *Client {
	return &Client{addr: addr}
}
