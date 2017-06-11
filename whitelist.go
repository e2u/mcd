package main

import (
	"bufio"
	"os"
	"strings"
	"sync"
)

const (
	WhiteListCacheSetKey = "WhiteListSet"
)

// 白名单数据结构
type WhiteList struct {
	lock sync.RWMutex
	m    map[string]bool
}

func NewWhiteList() *WhiteList {
	return &WhiteList{}
}

// LoadFromFile  从文件加载白名单
func (w *WhiteList) LoadFromFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)

	w.lock.Lock()
	defer w.lock.Unlock()
	w.m = make(map[string]bool)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		w.m[line] = true
	}

	return err
}

func (w *WhiteList) Exist(member string) bool {
	w.lock.RLock()
	defer w.lock.RUnlock()
	v, ok := w.m[member]
	return (v && ok)
}
