package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/e2u/goboot"
	"github.com/e2u/mcd/cache"
	"github.com/gorilla/mux"
)

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

	rs := preProcessRequestResources(strings.Split(r.FormValue("rc"), ","), func /*skip*/ (v string) bool {
		return !strings.HasSuffix(v, rcTypeExt)
	})
	goboot.Log.Debugf("request resouces: %v", rs)

	return rs
}

// 强制更新缓存
func (c *Controller) UpdateHandler(w http.ResponseWriter, r *http.Request) {

	rs := preProcessRequestResources(strings.Split(r.FormValue("rc"), ","), func /*skip*/ (v string) bool {
		return false
	})

	func(_rs []string) { // 这个不要起 gorouting 运行
		sort.Strings(_rs)
		Cache.Delete(strings.Join(rs, ",") + cachePNGSuffix)
	}(rs)

	scale := r.FormValue("scale")
	if scale == "" {
		scale = strconv.Itoa(DefaultScale)
	}
	rs = append(rs, fmt.Sprintf("$%s$", scale))
	sort.Strings(rs)

	orrs := strings.Join(rs, ",")
	go Cache.Delete(orrs)
	go Cache.Delete(orrs + cacheCSSSuffix)

	for _, r := range rs {
		if strings.HasPrefix(r, "$") { // 跳过图片比例
			continue
		}
		go Cache.Delete(r)
	}
	w.Write([]byte("OK"))
}

// 列出短服务器 tag
func (c *Controller) Tags(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(TL.List())
	if err != nil {
		io.Copy(w, strings.NewReader(err.Error()))
		return
	}
	io.Copy(w, bytes.NewReader(b))
}

// 整合 js,css 资源输出
// TODO: 优化一下,将实时读取独立资源整合输出,修改为双重缓存. (1)原始资源缓存;(2)整合后资源缓存
func (c *Controller) MergeHandler(w http.ResponseWriter, r *http.Request) {

	originRequestResources := processRequestResources(w, r)

	headerOutput := func(w http.ResponseWriter, ti time.Time) {
		w.Header().Set("Cache-Control", "max-age:1296000, public")
		w.Header().Set("Last-Modified", ti.Format(http.TimeFormat))
		w.Header().Set("Expires", ti.AddDate(0, 0, 20).Format(http.TimeFormat))
	}

	// 尝试取已经整合好的资源,如果有整合缓存,则直接输出
	orrs := strings.Join(originRequestResources, ",")
	if oc, err := Cache.Get(orrs); err == nil && oc != nil {
		goboot.Log.Debugf("merged cache: %v", orrs)
		headerOutput(w, oc.CreatedAt)
		io.Copy(w, bytes.NewReader(oc.Object))
		return
	}

	// 否则获取独立资源,整合输出并缓存,由于资源引用有顺序问题，暂时不考虑并发读取原始资源
	goboot.Log.Debugf("fetch origin resources: %v", originRequestResources)

	outputBytes := &bytes.Buffer{}
	for _, rc := range originRequestResources {
		if oc, err := getResource(rc, minifier); err == nil {
			ocInfo := fmt.Sprintf("/*\n * MCD Info:\n * Source: %s\n * CacheAt: %s\n * Length: %d Bytes\n * MD5Hash: %x\n */\n", oc.Source, oc.CreatedAt, oc.Length, oc.MD5Hash)
			outputBytes.WriteString(ocInfo)
			outputBytes.Write(oc.Object)
		}
	}

	headerOutput(w, time.Now())
	w.Write(outputBytes.Bytes())

	go Cache.Set(orrs, &cache.CacheObject{
		CreatedAt: time.Now(),
		Length:    uint64(outputBytes.Len()),
		MD5Hash:   md5.Sum(outputBytes.Bytes()),
		Object:    outputBytes.Bytes(),
		Source:    orrs,
	})

}
