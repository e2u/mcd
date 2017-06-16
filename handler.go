package main

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/e2u/goboot"
	"github.com/gorilla/mux"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"
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

// 处理请求的资源列表
func processRequestResources(w http.ResponseWriter, r *http.Request) []string {
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
		if v1 == "" || !strings.HasSuffix(v1, rcTypeExt) || isInArray(rs, v1) {
			continue
		}

		// 如果是在信任服务器中的资源
		if v2 := TL.GetByShort(v1); v2 != "" {
			rs = append(rs, v2)
			goboot.Log.Debugf("trust resource: %v", v2)
			continue
		}

		// 需要白名单验证
		if isInWhitelist(v1) {
			rs = append(rs, v1)
			goboot.Log.Debugf("white resource: %v", v1)
			continue
		}

	}

	goboot.Log.Debugf("request resouces: %v", rs)
	return rs
}

// 强制更新缓存
func (c *Controller) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	var rs []string
	for _, v := range strings.Split(r.FormValue("rc"), ",") {
		v1 := strings.TrimSpace(v)
		if v1 == "" || isInArray(rs, v1) {
			continue
		}

		if v2 := TL.GetByShort(v1); v2 != "" {
			rs = append(rs, v2)
			goboot.Log.Debugf("trust resource: %v", v2)
			continue
		}

		if isInWhitelist(v1) {
			rs = append(rs, v1)
			continue
		}
	}
	for _, r := range rs {
		go Cache.Delete(r)
	}
	w.Write([]byte("OK"))
}

// 处理 cors 访问控制
// 如果请求参数中有 cors 参数,则读取 cors 配置文件对请求做处理
// cors 配置文件也需要在白名单中
/*
func (c *Controller) cors(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if len(origin) == 0 {
		return
	}
	var varys []string
	cross := DefaultCross()
	func() {
		for _, o := range cross.AllowedOrigins {
			if origin == o {
				w.Header().Set("Access-Control-Allow-Origin", o)
				varys = append(varys, "Origin")
			}
			if o == "*" {
				w.Header().Set("Access-Control-Allow-Origin", o)
				return
			}
		}
	}()

	writeHeader := func(hs []string, header string) {
		var s []string
		for _, h := range hs {
			s = append(s, h)
		}
		w.Header().Set(header, strings.Join(s, ","))
	}

	writeHeader(cross.AllowedHeaders, "Access-Control-Allow-Headers")
	writeHeader(cross.AllowedMethods, "Access-Control-Allow-Methods")
	writeHeader(varys, "Vary")
}
*/

// 整合资源输出
func (c *Controller) MergeHandler(w http.ResponseWriter, r *http.Request) {

	var outputResources []*cache.CacheObject
	var etagSource []byte
	for _, rc := range processRequestResources(w, r) {
		if oc, err := getResource(rc); err == nil {
			outputResources = append(outputResources, oc)
			etagSource = append(etagSource, oc.MD5Hash[:]...)
		}
	}
	// 如果客户端有 If-None-Match ,且所传递的值和 etag 相同,则返回 304
	etag := fmt.Sprintf("%x", md5.Sum(etagSource))
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("ETag", etag)
	for _, oc := range outputResources {
		ocInfo := fmt.Sprintf("/*\n * MCD Info:\n * Source: %s\n * CacheAt: %s\n * Length: %d Bytes\n * MD5Hash: %x\n */\n", oc.Source, oc.CreatedAt, oc.Length, oc.MD5Hash)
		io.Copy(w, strings.NewReader(ocInfo))
		io.Copy(w, bytes.NewReader(oc.Object))
	}
}

// 尝试从本地缓存或远端获取资源
func getResource(url string) (*cache.CacheObject, error) {

	if oc, err := Cache.Get(url); err == nil && oc != nil {
		return oc, err
	}

	if b, err := HttpGet(url); err == nil {

		r := bytes.NewBuffer(b)
		w := &bytes.Buffer{}

		if err := minifier(w, r, url); err != nil {
			goboot.Log.Errorf("minifiter error: %v", err.Error())
			return nil, err
		}

		oc := &cache.CacheObject{
			CreatedAt: time.Now(),
			Length:    uint64(len(w.Bytes())),
			MD5Hash:   md5.Sum(w.Bytes()),
			Object:    w.Bytes(),
			Source:    url,
		}
		Cache.Set(url, oc)
		return oc, err
	} else {
		return nil, err
	}
}

// 对资源进行压缩
func minifier(w io.Writer, r io.Reader, url string) error {
	m := minify.New()
	switch {
	case strings.HasSuffix(url, RcTypeJSExt):
		return js.Minify(m, w, r, nil)
	case strings.HasSuffix(url, RcTypeCSSExt):
		return css.Minify(m, w, r, nil)
	default:
		return errors.New("unknow type")
	}
	return errors.New("unknow type")
}
