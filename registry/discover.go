package registry

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
)

// EtcdDial 使用etcd解析器创建一个gRPC客户端连接，该连接将连接到名为service的服务。
func EtcdDial(c *clientv3.Client, service string) (*grpc.ClientConn, error) {
	etcdResolver, err := resolver.NewBuilder(c)
	if err != nil {
		return nil, err
	}
	return grpc.Dial("etcd:///"+service, grpc.WithResolvers(etcdResolver), grpc.WithInsecure())
}
