package singleflight

import (
	"sync"
)

// singleflight模块提供防止缓存击穿的能力
// 当有多个请求并发访问相同的缓存时
// 只会有一个请求能够申请到Group.mu
// 从而防止了大量请求同时访问

// call 表示正在进行中的请求
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group 管理不同key的请求
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	// 对于一个key而言 Group中已经有了相同的请求
	// 说明此时应该等待已经发起的请求做完
	// 因此进入wait状态
	// 此时可以将锁释放 这样其他相同的请求也能直接进入到等待状态
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	// 如果Group中没有相同的请求
	// 说明当前请求是第一个获取到资源的请求
	// 那么它应该继续完成后续的动作
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
