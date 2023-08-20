package registry

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"log"
	"time"
)

// registry 提供etcd服务注册与服务发现的能力

var (
	defaultEtcdConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	}
)

// etcdAdd 添加一对kv到etcd
func etcdAdd(c *clientv3.Client, lid clientv3.LeaseID, service string, addr string) error {
	em, err := endpoints.NewManager(c, service)
	if err != nil {
		return err
	}
	return em.AddEndpoint(c.Ctx(), service+"/"+addr, endpoints.Endpoint{Addr: addr}, clientv3.WithLease(lid))
}

// Registry 注册一个服务到etcd
func Registry(service, addr string, stop chan error) error {
	// 创建etcd client
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		return fmt.Errorf("create etcd client failed: %v", err)
	}
	defer cli.Close()

	// 创建一个租约
	resp, err := cli.Grant(context.Background(), 5)
	if err != nil {
		return fmt.Errorf("create etcd lease failed: %v", err)
	}
	leaseid := resp.ID

	// 服务注册
	err = etcdAdd(cli, leaseid, service, addr)
	if err != nil {
		return fmt.Errorf("add etcd failed: %v", err)
	}

	// 设置服务心跳检测
	ch, err := cli.KeepAlive(context.Background(), leaseid)
	if err != nil {
		return fmt.Errorf("set keepalive failed: %v", err)
	}

	log.Println(fmt.Sprintf("[%s] register service ok\n", addr))
	for {
		select {
		case err := <-stop:
			// 监听服务本身报错
			if err != nil {
				log.Println(err)
			}
			return err
		case <-cli.Ctx().Done():
			// 服务结束
			log.Println("service closed")
		case _, ok := <-ch:
			// keep alive 失效
			if !ok {
				log.Println("keep alive lose")
				_, err := cli.Revoke(context.Background(), leaseid)
				return err
			}
		}
	}
}
