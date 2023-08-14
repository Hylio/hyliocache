package consistenthash

import (
	"strconv"
	"testing"
)

func TestHash(t *testing.T) {
	hashfn := func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	}

	m := New(3, hashfn)
	m.Add("2", "5", "8") // keys = [2 5 8 12 15 18 22 25 28]
	testcases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "5",
		"27": "8",
	}

	for k, v := range testcases {
		if m.Get(k) != v {
			t.Errorf("hashing %s, want %s, but got %s", k, v, m.Get(k))
		}
	}

	m.Add("3") // keys = [2,3,5,8,12,13,15,18,22,23,25,28]
	testcases["23"] = "3"

	for k, v := range testcases {
		if m.Get(k) != v {
			t.Errorf("hashing %s, want %s, but got %s", k, v, m.Get(k))
		}
	}
}
