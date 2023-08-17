package hyliocache

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	pb "github.com/hylio/hyliocache/hyliocachepb"
	"io"
	"net/http"
	"net/url"
)

// client 实现了访问其他远程节点并获取缓存的能力

type Client struct {
	addr string // 定义将要访问的服务的地址 ip:port
}

func (h *Client) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf("%v%v/%v", h.addr, url.QueryEscape(in.GetGroup()), url.QueryEscape(in.GetKey()))
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Server returned： %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}
	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}

func NewClient(addr string) *Client {
	return &Client{addr: addr}
}
