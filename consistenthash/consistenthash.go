package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash  定义哈希函数
type Hash func(data []byte) uint32

type Map struct {
	hash     Hash           // 哈希函数
	replicas int            // 虚拟节点倍数
	keys     []int          // 哈希环
	hashMap  map[int]string // 节点与虚拟节点的对应关系
}

func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			//计算节点哈希值
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	idx := sort.SearchInts(m.keys, hash)
	return m.hashMap[m.keys[idx%len(m.keys)]]
}

func New(replicas int, hash Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     hash,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		// 默认哈希函数
		m.hash = crc32.ChecksumIEEE
	}
	return m
}
