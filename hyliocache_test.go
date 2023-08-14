package hyliocache

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed")
	}
}

var db = map[string]string{
	"zhanghao":    "hylio",
	"wangrui":     "civet",
	"zhouruqiang": "dio",
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	c := NewGroup("eng_name", 2<<10, GetterFunc(func(key string) ([]byte, error) {
		if v, ok := db[key]; ok {
			if _, ok := loadCounts[key]; !ok {
				loadCounts[key] = 0
			}
			loadCounts[key]++
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))

	for k, v := range db {
		// 第一次访问 调用回调函数 访问db
		if view, err := c.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value")
		}
		// 第二次访问 直接命中缓存
		if _, err := c.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	if view, err := c.Get("unknown"); err == nil {
		t.Fatalf("should be empty, but got %s", view)
	}
}
