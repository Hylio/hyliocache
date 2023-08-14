package hyliocache

/*
ByteView是对byte数组的只读封装
*/

type ByteView struct {
	// 为什么使用byte数组
	// 为了保证可以覆盖string/图片等各种格式
	b []byte
}

func (b ByteView) Len() int {
	return len(b.b)
}

// ByteSlice 返回数据的的深拷贝副本 保证byteview的只读
func (b ByteView) ByteSlice() []byte {
	return cloneBytes(b.b)
}

func (b ByteView) String() string {
	return string(b.b)
}

func cloneBytes(b []byte) []byte {
	clone := make([]byte, len(b))
	copy(clone, b)
	return clone
}
