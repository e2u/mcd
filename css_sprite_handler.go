// 将多个图片整合成一张大图,然后输出 css 和图片
// 127.0.0.1:9000/scss?rc=i1:ShimoIcon.png,i1:icon_back@2x.png,i1:icon_circle.png,i1:icon_closed.png,i1:icon_downblack.png,i1:icon_liuliang@2x.png,i1:icon_right.png,i1:icon_select.png,i1:check@2x.png,i1:icon_chongzhi@2x.png,i1:icon_circle2.png,i1:icon_down.png,i1:icon_list.png,i1:icon_record@2x.png,i1:icon_search@2x.png,i1:icon_up.png
package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/e2u/goboot"
	"github.com/e2u/mcd/cache"
)

var (
	csssReType = []string{".png", ".jpg", ".jpeg", ".gif"} // 支持的原始图片扩展名
	// 避免 key 相同生成的缓存冲突
	cacheCSSSuffix = ",css"
	cachePNGSuffix = ",png"
)

// 生成 scss
func (c *Controller) CSSSpriteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")

	reqRc := r.FormValue("rc")
	rs := preProcessRequestResources(strings.Split(reqRc, ","), func /*skip*/ (v string) bool {
		return !isInArray(csssReType, filepath.Ext(v))
	})
	sort.Strings(rs)

	headerOutput := func(w http.ResponseWriter, ti time.Time) {
		w.Header().Set("Cache-Control", "max-age:1296000, public")
		w.Header().Set("Last-Modified", ti.Format(http.TimeFormat))
		w.Header().Set("Expires", ti.AddDate(0, 0, 20).Format(http.TimeFormat))
	}

	// 尝试读取已经生成的 css 输出
	orrs := strings.Join(rs, ",")
	if oc, err := Cache.Get(orrs + cacheCSSSuffix); err == nil && oc != nil {
		goboot.Log.Debugf("merged cache: %v", orrs)
		headerOutput(w, oc.CreatedAt)
		io.Copy(w, bytes.NewReader(oc.Object))
		return
	}

	var sis []*SpriteImage

	for _, rc := range rs {
		oc, err := getResource(rc)
		if err != nil {
			goboot.Log.Errorf("get resources error: %v", err.Error())
			continue
		}
		i, format, err := image.Decode(bytes.NewReader(oc.Object))
		if err != nil {
			goboot.Log.Errorf("decode image %v error: %v", rc, err.Error())
			continue
		}
		sis = append(sis, &SpriteImage{
			Format:   format,
			Image:    i,
			Width:    i.Bounds().Max.X,
			Height:   i.Bounds().Max.Y,
			FileName: rc,
		})
	}

	iw, ih := calculateImageDimension(sis)
	simg := image.NewRGBA64(image.Rect(0, 0, iw, ih))

	outBuf := new(bytes.Buffer)
	var nextY int

	spriteImageFullUrl := func() string {
		siteBase := goboot.Config.MustString("site.base", "http://127.0.0.1:9000")
		return fmt.Sprintf("%s/scss-image?rc=%s", siteBase, reqRc)
	}()

	var outCSS []string
	outCSS = append(outCSS, fmt.Sprintf(".mcd-scss {background: url('%s') no-repeat top left;background-size: %srem auto;}", spriteImageFullUrl, px2rem(iw)))
	for _, si := range sis {
		draw.Draw(simg, simg.Bounds().Add(image.Point{0, nextY}), si.Image, image.Point{0, 0}, draw.Src)
		outCSS = append(outCSS, genCss(si, nextY, spriteImageFullUrl))
		nextY += si.Height
	}

	outCSSStr := strings.Join(outCSS, "\n")

	if err := png.Encode(outBuf, simg); err == nil {
		Cache.Set(orrs+cachePNGSuffix, &cache.CacheObject{
			CreatedAt: time.Now(),
			Length:    uint64(outBuf.Len()),
			MD5Hash:   md5.Sum(outBuf.Bytes()),
			Object:    outBuf.Bytes(),
			Source:    orrs,
		})
	}

	Cache.Set(orrs+cacheCSSSuffix, &cache.CacheObject{
		CreatedAt: time.Now(),
		Length:    uint64(len(outCSSStr)),
		MD5Hash:   md5.Sum([]byte(outCSSStr)),
		Object:    []byte(outCSSStr),
		Source:    orrs,
	})

	io.Copy(w, strings.NewReader(outCSSStr))

}

// 生成 scss 引用的图片
func (c *Controller) CSSSpriteImageHandler(w http.ResponseWriter, r *http.Request) {
	rs := preProcessRequestResources(strings.Split(r.FormValue("rc"), ","), func /*skip*/ (v string) bool {
		return !isInArray(csssReType, filepath.Ext(v))
	})
	sort.Strings(rs)

	headerOutput := func(w http.ResponseWriter, ti time.Time) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "max-age:1296000, public")
		w.Header().Set("Last-Modified", ti.Format(http.TimeFormat))
		w.Header().Set("Expires", ti.AddDate(0, 0, 20).Format(http.TimeFormat))
	}

	// 尝试读取整合大图输出
	orrs := strings.Join(rs, ",")
	if oc, err := Cache.Get(orrs + cachePNGSuffix); err == nil && oc != nil {
		goboot.Log.Debugf("merged cache: %v", orrs)
		headerOutput(w, oc.CreatedAt)
		io.Copy(w, bytes.NewReader(oc.Object))
		return
	}

	// 否则输出 204 无内容
	w.WriteHeader(http.StatusNoContent)
}

// 生成 scss 预览页面,方便前端开发选用图片
func (c *Controller) CSSSpriteReviewHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html;charset=utf-8")
}

// 计算出合并一组图片所需的尺寸,生产合并图片所需的底图大小
// 返回: width,height
func calculateImageDimension(sis []*SpriteImage) (int, int) {
	var maxWidth, sumHeight int
	for _, si := range sis {
		if si.Width >= maxWidth {
			maxWidth = si.Width
		}
		sumHeight += si.Height
	}
	return maxWidth, sumHeight
}

// 生成 css
func genCss(si *SpriteImage, nextY int, fullUrl string) string {
	fn := func() string {
		s := filepath.Base(si.FileName)
		for _, b := range cssKeyWork {
			s = strings.Replace(s, string(b), "_", -1)
		}
		return string(s)
	}()
	return fmt.Sprintf(".%s{width: %srem; height: %srem; background-position: 0 -%srem;}", fn, px2rem(si.Width), px2rem(si.Height), px2rem(nextY))
}

// px 转 rem
func px2rem(px int) string {
	return fmt.Sprintf("%.2f", float64(px)*0.01)
}
