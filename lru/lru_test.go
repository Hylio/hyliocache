package lru

import (
	"fmt"
	"reflect"
	"testing"
)

type Str string

func (s Str) Len() int {
	return len(s)
}

func TestCache_Get(t *testing.T) {
	lruCache := New(0, nil)
	lruCache.Add("name", Str("hylio"))
	if v, ok := lruCache.Get("name"); !ok || string(v.(Str)) != "hylio" {
		t.Error("hit fail")
	}
	if _, ok := lruCache.Get("age"); ok {
		t.Error("miss fail")
	}
}

func TestCache_Remove(t *testing.T) {
	k1, k2, k3 := "name", "age", "sex"
	v1, v2, v3 := "hylio", "24", "male"
	cap := len(k1 + k2 + v1 + v2)
	lruCache := New(int64(cap), nil)
	lruCache.Add(k1, Str(v1))
	lruCache.Add(k2, Str(v2))
	lruCache.Add(k3, Str(v3))
	if _, ok := lruCache.Get(k3); !ok {
		t.Error("hit k3 fail")
	}
	fmt.Print(lruCache.cache)
	if _, ok := lruCache.Get(k1); ok || lruCache.Len() != 2 {
		t.Error("remove k1 fail")
	}
}

func TestCache_Add(t *testing.T) {
	k1 := "name"
	v1 := "hylio"
	v2 := "zh"
	cap := len(k1 + v1)
	lruCache := New(int64(cap), nil)
	lruCache.Add(k1, Str(v1))
	if _, ok := lruCache.cache[k1]; !ok || lruCache.nbytes != int64(cap) {
		t.Error("hit k1 fail")
	}
	lruCache.Add(k1, Str(v2))
	if v, ok := lruCache.Get(k1); !ok || string(v.(Str)) != v2 {
		t.Error("update k1 fail")
	}
}

func TestOnEvicted(t *testing.T) {
	del := make([]string, 0)
	callback := func(key string, value Value) {
		del = append(del, key)
	}
	k1, k2, k3 := "name", "age", "sex"
	v1, v2, v3 := "hylio", "24", "male"
	cap := len(k1 + k2 + v1 + v2)
	lruCache := New(int64(cap), callback)
	lruCache.Add(k1, Str(v1))
	lruCache.Add(k2, Str(v2))
	lruCache.Add(k3, Str(v3))
	expect := []string{k1}

	if !reflect.DeepEqual(expect, del) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}

func TestLimit(t *testing.T) {
	k1, k2, k3 := "name", "age", "sex"
	v1, v2, v3 := "hylio", "24", "malemalemalemale"
	cap := 10
	lruCache := New(int64(cap), nil)
	lruCache.Add(k1, Str(v1))
	lruCache.Add(k2, Str(v2))
	lruCache.Add(k3, Str(v3))
	if _, ok := lruCache.Get(k3); ok || lruCache.Len() != 0 {
		t.Error("limit fail")
	}
}
