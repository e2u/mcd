// 处理 cross 配置
package main

var (
	defaultCorsOrigins = []string{"*"}
	defaultCorsMethods = []string{"GET", "HEAD", "POST"}
	defaultCorsHeaders = []string{"Accept", "Accept-Language", "Content-Language", "Origin"}
)

const (
	corsOptionMethod           string = "OPTIONS"
	corsAllowOriginHeader      string = "Access-Control-Allow-Origin"
	corsExposeHeadersHeader    string = "Access-Control-Expose-Headers"
	corsMaxAgeHeader           string = "Access-Control-Max-Age"
	corsAllowMethodsHeader     string = "Access-Control-Allow-Methods"
	corsAllowHeadersHeader     string = "Access-Control-Allow-Headers"
	corsAllowCredentialsHeader string = "Access-Control-Allow-Credentials"
	corsRequestMethodHeader    string = "Access-Control-Request-Method"
	corsRequestHeadersHeader   string = "Access-Control-Request-Headers"
	corsOriginHeader           string = "Origin"
	corsVaryHeader             string = "Vary"
	corsOriginMatchAll         string = "*"
)

// 解析传入的配置文件
type Cross struct {
	AllowedHeaders   []string
	AllowedMethods   []string
	AllowedOrigins   []string
	ExposedHeaders   []string
	MaxAge           int
	IgnoreOptions    bool
	AllowCredentials bool
}

func DefaultCross() *Cross {
	return &Cross{
		AllowedOrigins: defaultCorsOrigins,
		AllowedMethods: defaultCorsMethods,
		AllowedHeaders: defaultCorsHeaders,
	}
}
