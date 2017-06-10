package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/weidewang/mcd/cache"
)

// 指定的字符串是否在数组中
func isInArray(as []string, s string) bool {
	for _, a := range as {
		if a == s {
			return true
		}
	}
	return false
}

// 指定的字符串是否在白名单中
func isInWhitelist(s string) bool {
	return WL.Exist(s)
}

const (
	RcTypeJSExt  = ".js"
	RcTypeCSSExt = ".css"
)

func cleanRequestResources(w http.ResponseWriter, r *http.Request) []string {
	vars := mux.Vars(r)
	rcType := vars["rcType"]

	var rcTypeExt string

	switch rcType {
	case "js":
		w.Header().Set("Content-Type", "application/javascript")
		rcTypeExt = RcTypeJSExt
	case "css":
		w.Header().Set("Content-Type", "text/css")
		rcTypeExt = RcTypeCSSExt
	}

	var rs []string
	for _, v := range strings.Split(r.FormValue("rc"), ",") {
		v1 := strings.TrimSpace(v)
		if len(v1) == 0 || !strings.HasSuffix(v1, rcTypeExt) || isInArray(rs, v1) || !isInWhitelist(v1) {
			continue
		}
		rs = append(rs, v1)
	}

	return rs
}

// 强制更新缓存
func (c *Controller) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	var rs []string
	for _, v := range strings.Split(r.FormValue("rc"), ",") {
		v1 := strings.TrimSpace(v)
		if len(v1) == 0 || isInArray(rs, v1) || !isInWhitelist(v1) {
			continue
		}
		rs = append(rs, v1)
	}
	for _, r := range rs {
		go Cache.Delete(r)
	}
	w.Write([]byte("OK"))
}

// 整合资源输出
func (c *Controller) MergeHandler(w http.ResponseWriter, r *http.Request) {

	var outputResources []*cache.CacheObject
	var etagSource []byte
	for _, rc := range cleanRequestResources(w, r) {
		if oc, err := getResource(rc); err == nil {
			outputResources = append(outputResources, oc)
			etagSource = append(etagSource, oc.MD5Hash[:]...)
		}
	}
	// 如果客户端有 If-None-Match ,且所传递的值和 etag 相同,则返回 304
	etag := fmt.Sprintf("%x", md5.Sum(etagSource))
	w.Header().Set("ETag", etag)
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	for _, oc := range outputResources {
		ocInfo := fmt.Sprintf("/*\n * MCD Info:\n * Source: %s\n * CacheAt: %s\n * Length: %d Bytes\n * MD5Hash: %x\n */\n", oc.Source, oc.CreatedAt, oc.Length, oc.MD5Hash)
		io.Copy(w, strings.NewReader(ocInfo))
		io.Copy(w, bytes.NewReader(oc.Object))
	}
}

func getResource(url string) (*cache.CacheObject, error) {

	if oc, err := Cache.Get(url); err == nil && oc != nil {
		return oc, err
	}

	if b, err := HttpGet(url); err == nil {
		oc := &cache.CacheObject{
			CreatedAt: time.Now(),
			Length:    uint64(len(b)),
			MD5Hash:   md5.Sum(b),
			Object:    b,
			Source:    url,
		}
		Cache.Set(url, oc)
		return oc, err
	} else {
		return nil, err
	}

}
