package main

import "net/http"

// 输出默认头信息
func headerMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Set("X-Server", "MCD/"+Version)
	w.Header().Set("Connection", "keep-alive")
	next(w, r)
}

func corsRequestMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Header.Get("Origin") == "" {
		next(w, r)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "ETag")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,If-None-Match,Cache-Control,Content-Type,ETag")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Max-Age", "1728000")
		w.Header().Set("Content-Type", "text/plain charset=UTF-8")
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Add("Vary", "Origin")

	next(w, r)
}
