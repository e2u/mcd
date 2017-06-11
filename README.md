# MDC


资源整合服务，可将多个分散在不同服务器，不同访问协议的 js,css 文件做打包输出并做缓存，用于解决浏览器引用大量 js,css 文件时过多的网络连接。


##  白名单

所有需要做整合输出的资源都需要在白名单中明确列出，白名单不支持正则表达，区分大小写，不在白名单中的原始资源不会做处理。

修改白名单后需要重启服务或等待10秒自动重新加载


## 整合输出接口

 本接口用于整合多个远程资源做输出:

`[GET|POST|HEAD] /[js|css]?rc=<resources.....>`

接口只有一个参数: rc ,参数值为原始资源的完整 url ,多个资源用逗号做分隔,例如:


`http[s]://domain.com/js?rc=http://src1.com/a.js,https://src1.com/b.js`


## 强制更新缓存接口

本服务不会主动检查原始资源的更新，如已缓存的原始资源发生了变化，需要调用本接口做强制更新

`[GET|POST|HEAD] /update?rc=<resources.....>`

强制更新接口可以同时更新多个不同类型的资源

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