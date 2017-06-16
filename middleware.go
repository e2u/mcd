package main

import "net/http"

// 输出默认头信息
func headerMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Set("X-Server", "MCD/"+Version)
	w.Header().Set("Connection", "keep-alive")
	next(w, r)
}

func corsRequestMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	next(w, r)
}
