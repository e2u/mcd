package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

var (
	// 路径清洗正则
	prefixRegexp = regexp.MustCompile(`^[a-zA-Z0-9_]`)
	dotRegexp    = regexp.MustCompile(`\.{2,}`)
	slashRegexp  = regexp.MustCompile(`/{2,}`)
)

// 信任服务器列表,用于缩短源文件的访问路径
type TrustServer struct {
	sync.RWMutex
	m map[string]*url.URL
}

func NewTrustServer() *TrustServer {
	return &TrustServer{
		m: make(map[string]*url.URL),
	}
}

func (t *TrustServer) Set(tag string, u *url.URL) {
	t.Lock()
	defer t.Unlock()
	t.m[strings.TrimSpace(tag)] = u
}

func (t *TrustServer) Get(tag string) *url.URL {
	t.RLock()
	defer t.RUnlock()
	if v, ok := t.m[tag]; ok {
		return v
	}
	return nil
}

// 根据短请求地址返回完整的资源地址,短请求地址样式:
// s1:/jquery-latest.js  或 s1:/jquery-3.2.1.min.js  或 s2:/2.0/typo.css
func (t *TrustServer) GetByShort(su string) string {
	tag, path := func() (string, string) {
		as := strings.Split(su, ":")
		return as[0], as[1]
	}()
	u := t.Get(tag)
	if u == nil {
		return ""
	}
	pu := u.String()
	if strings.HasSuffix(pu, "/") {
		pu = pu[:len(pu)-1]
	}

	// 做路径清洗
	path = dotRegexp.ReplaceAllString(path, ".")
	path = slashRegexp.ReplaceAllString(path, "/")
	for !prefixRegexp.MatchString(path) {
		path = path[1:]
	}

	if u, err := url.Parse(fmt.Sprintf("%s/%s", pu, path)); err == nil {
		return u.String()
	}

	return ""
}

func (t *TrustServer) List() map[string]string {
	t.RLock()
	defer t.RUnlock()
	r := make(map[string]string)
	for k, v := range t.m {
		r[k] = v.String()
	}
	return r
}
