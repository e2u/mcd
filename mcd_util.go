package main

import (
	"bytes"
	"crypto/md5"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/e2u/goboot"
	"github.com/e2u/mcd/cache"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"
)

var (
	// 路径清洗正则
	prefixRegexp = regexp.MustCompile(`^[a-zA-Z0-9_]`)
	dotRegexp    = regexp.MustCompile(`\.{2,}`)
	slashRegexp  = regexp.MustCompile(`/{2,}`)
)

// 做路径清洗
func cleanPrefixPath(path string) string {
	path = dotRegexp.ReplaceAllString(path, ".")
	path = slashRegexp.ReplaceAllString(path, "/")
	for !prefixRegexp.MatchString(path) {
		path = path[1:]
	}
	return path
}

// 尝试从本地缓存或远端获取资源
// afterFuncs 可对获取到的原始资源做处理
func getResource(url string, afterFuncs ...func(w io.Writer, r io.Reader, url string) error) (*cache.CacheObject, error) {
	if oc, err := Cache.Get(url); err == nil && oc != nil {
		return oc, err
	}

	if b, err := HttpGet(url); err == nil {
		r := bytes.NewBuffer(b)
		w := &bytes.Buffer{}

		if len(afterFuncs) == 0 {
			io.Copy(w, r)
		} else {
			for _, afterFunc := range afterFuncs {
				if err := afterFunc(w, r, url); err != nil {
					goboot.Log.Errorf("exec afterFunc error: %v", err.Error())
					return nil, err
				}
			}
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

func noneMinifier(w io.Writer, r io.Reader, url string) error {
	io.Copy(w, r)
	return nil
}

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

// 预处理请求资源
// rcs 请求的原始资源
// skip 要跳过的资源方法,默认会跳过空字符串,重复的资源
func preProcessRequestResources(rcs []string, skipFunc func(rc string) bool) []string {
	var rs []string
	for _, v := range rcs {
		v1 := strings.TrimSpace(v)
		if v1 == "" || isInArray(rs, v1) || skipFunc(v1) {
			continue
		}
		// 如果是在信任服务器中的资源
		if v2 := TL.GetByShort(v1); v2 != "" {
			if isInArray(rs, v2) {
				continue
			}
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
	return rs
}
