package hyliocache

import pb "github.com/hylio/hyliocache/hyliocachepb"

// PeerPicker 保证了获取远端分布式节点的能力
type PeerPicker interface {
	// PickPeer 根据key选择远端节点
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 保证了可以获取缓存的能力
type PeerGetter interface {
	//Get(group string, key string) ([]byte, error)
	Get(in *pb.Request, out *pb.Response) error
}
