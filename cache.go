package hyliocache

/*
cache 模块负责提供缓存的并发控制
*/

import (
	"hyliocache/lru"
	"sync"
)

// 把算法和实际缓存进行了分离
// 这样如果要更换算法 就只需要替换lru

type cache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64 // 最大缓存容量
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		// 懒初始化  在第一次使用时再初始化
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
