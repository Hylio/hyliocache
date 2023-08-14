package lru

import "container/list"

/*
	使用双向链表和字典结合的LRU
	双向链表可以保证O(1)的变更到最前/最后
	字典可以保证O(1)的查询
*/

type Cache struct {
	maxBytes  int64                         // 最大内存限制
	nbytes    int64                         // 当前已使用内存
	ll        *list.List                    // 双向链表
	cache     map[string]*list.Element      // map
	OnEvicted func(key string, value Value) // 回调函数 在淘汰数据时执行其他操作
}

type entry struct {
	key   string
	value Value
}

// Value 为了通用性 定义为interface 方法只包含Len 用于返回值所占的内存大小
type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		// 最近访问过  放到队尾
		c.ll.MoveToBack(ele)
		kv := ele.Value.(*entry)
		value = kv.value
		return value, true
	}
	return
}

// Remove 删除掉最近最久未访问 即链表最前面的元素
func (c *Cache) Remove() {
	ele := c.ll.Front()
	// 不加if nil 的判断的话  如果删完就会报错
	if ele != nil {
		kv := ele.Value.(*entry)
		key, value := kv.key, kv.value
		delete(c.cache, key)
		// 计算内存时需要考虑字典占用的内存
		c.nbytes -= int64(value.Len()) + int64(len(key))
		if c.OnEvicted != nil {
			c.OnEvicted(key, value)
		}
		c.ll.Remove(ele)
	}
}

// Add 添加或修改元素
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		// update
		c.ll.MoveToBack(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		//for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		//	c.Remove()
		//}
		kv.value = value
	} else {
		// add
		c.nbytes += int64(value.Len()) + int64(len(key))
		//for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		//	c.Remove()
		//}
		ele := c.ll.PushBack(&entry{key: key, value: value})
		c.cache[key] = ele
	}
	// TODO: 会先超内存再删
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.Remove()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
