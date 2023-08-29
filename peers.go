package hyliocache

import pb "github.com/hylio/hyliocache/hyliocachepb"

// PeerPicker 保证了获取远端分布式节点的能力 用Server实现了这个接口
type PeerPicker interface {
	// PickPeer 根据key选择远端节点
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 保证了可以获取缓存的能力 用Client实现了这个接口
type PeerGetter interface {
	Get(in *pb.Request) ([]byte, error)
}
