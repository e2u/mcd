package cache

import (
	"encoding/gob"
	"io"
	"time"
)

type CacheObject struct {
	Source    string    // 源地址
	Length    uint64    // 长度
	MD5Hash   [16]byte  // 内容 hash 值
	CreatedAt time.Time // 缓存创建的时间
	Object    []byte    // 缓存的内容
}

func LoadCacheObject(r io.Reader) (*CacheObject, error) {
	var co CacheObject
	err := gob.NewDecoder(r).Decode(&co)
	return &co, err
}

func DumpCacheObject(w io.Writer, oc *CacheObject) error {
	return gob.NewEncoder(w).Encode(oc)
}
