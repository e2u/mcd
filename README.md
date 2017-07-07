# MCD


资源整合服务，可将多个分散在不同服务器，不同访问协议的 js,css,image 文件做打包输出并做缓存，用于解决浏览器引用大量 js,css,image 文件时过多的网络连接。


##  白名单

所有需要做整合输出的资源都需要在白名单中明确列出，白名单不支持正则表达，区分大小写，不在白名单中的原始资源不会做处理。

修改白名单后需要重启服务或等待10秒自动重新加载

## 信任服务器

可在配置文件中添加一个或多个信任服务器列表,存储在信任服务器中的源文件不需要一一添加白名单，同时可缩短整合请求的地址。


信任服务器列表,可填写多个,逗号","分隔,格式为: `<tag1>=<url>`,`<tag2>=<url2>`,....

在接口访问时可缩短源文件的路径

例如: `trust.server= s1=http://http://code.jquery.com/,s2=http://apps.bdimg.com/libs/typo.css/`

在访问时,可以使用
`http[s]://domain.com/js?rc=s1:/jquery-latest.js,s1:/jquery-3.2.1.min.js` 这样的形式进行访问
`http[s]://domain.com/css?rc=s2:/2.0/typo.css`


## 整合输出接口

 本接口用于整合多个远程资源做输出:

`[GET|POST|HEAD] /[js|css]?rc=<resources.....>`

接口只有一个参数: rc ,参数值为原始资源的完整 url ,多个资源用逗号做分隔,例如:


`http[s]://domain.com/js?rc=http://src1.com/a.js,https://src1.com/b.js`


## 图片整合输出 CSS Sprite 接口

接收多个源图片参数，并将图片合并成一个大的图片输出，同时输出 css 样式表,可以直接使用。

`[GET] /scss?rc=<resources....>&scale=100`  输出 css 样式

**参数:**
	* rs: 需要整合的资源列表
	* scale: 生成css的缩放比例,默认100,生成的 css 中图片的尺寸(rem)根据输入的 scale 参数做响应的缩放处理。

使用范例:

```
<!DOCTYPE html>
<html>
    <head>
		<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0,minimum-scale=1.0,user-scalable=0">
        <link href="http://127.0.0.1:9000/scss?r=aaaa&rc=i1:ShimoIcon.png,i1:icon_back@2x.png,i1:icon_circle.png,i1:icon_closed.png,i1:icon_downblack.png,i1:icon_liuliang@2x.png,i1:icon_right.png,i1:icon_select.png,i1:check@2x.png,i1:icon_chongzhi@2x.png,i1:icon_circle2.png,i1:icon_down.png,i1:icon_list.png,i1:icon_record@2x.png,i1:icon_search@2x.png,i1:icon_up.png&scale=60" rel="stylesheet" type="text/css"/>
				<script type="text/javascript" src="rem_750.js"></script>
    </head>
    <body style="background-color:#FFFFFF">
			hello
        <div class="mcd-scss ShimoIcon_png"></div>
        <div class="mcd-scss check_2x_png"></div>
        <div class="mcd-scss icon_back_2x_png"></div>
        <div class="mcd-scss icon_chongzhi_2x_png"></div>
        <div class="mcd-scss icon_circle_png"></div>
        <div class="mcd-scss icon_record_2x_png"></div>
    </body>
</html>
```

**rem_750.js"
```
/* fix the code flash the page  */
var globalWidth = document.documentElement.clientWidth;//window.innerWidth || document.documentElement.clientWidth || document.body.clientWidth;
var radixNO = 100/750*globalWidth;
document.documentElement.style.fontSize = radixNO + 'px';
/* fit document fit the screen, get the radix */
(function (doc, win) {
    var docEl = doc.documentElement,
            resizeEvt = 'orientationchange' in window ? 'orientationchange' : 'resize',
            recalc = function () {
                var globalWidth = window.innerWidth;// for judge the screen ??
                var clientWidth = docEl.clientWidth;
                if (!clientWidth) return;
                docEl.style.fontSize = 100 * (clientWidth / 750) + 'px';
            };
    if (!doc.addEventListener) return;
    win.addEventListener(resizeEvt, recalc, false);
    doc.addEventListener('DOMContentLoaded', recalc, false);
})(document, window);

```




## 强制更新缓存接口

本服务不会主动检查原始资源的更新，如已缓存的原始资源发生了变化，需要调用本接口做强制更新

`[GET|POST|HEAD] /update?rc=<resources.....>&scale=100`

强制更新接口可以同时更新多个不同类型的资源

## 信任服务器列表

`[GET] /tags`

列出已配置的信任服务器列表

## 限制

* 支持的原始资源类型: js,css
* 原始资源的扩展名要求全部是小写字母,例如: a.js,a.css 是合法的名字, a.Js,a.Css 都是不合法名字


## 计划

* js,css 压缩

## 运行方式

### 配置文件

见 conf/app.conf

### Docker

`docker build -t mcd-docker .`
`docker run mcd-docker`

### 直接运行

`go build -o mcd *.go`
`./mcd`
