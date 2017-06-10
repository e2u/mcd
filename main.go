package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/e2u/goboot"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"github.com/weidewang/mcd/cache"
)

const (
	Version = "1.0.0"
)

var (
	ListenPort int
	RunEnv     string
	WL         *WhiteList
	WLModTime  time.Time
	Cache      cache.Storager
)

type Controller struct {
}

// 初始化缓存
func initCache() error {
	Cache = cache.NewFileStorage(goboot.Config.MustString("file.storage.path", "/tmp/mcd"))
	return nil
}

// 初始化白名单
func initWhitelist() error {
	WL = NewWhiteList()
	wlfile := goboot.Config.MustString("resources.whitelist")
	go func() {
		for {
			if s, err := os.Stat(wlfile); err == nil && !s.ModTime().Equal(WLModTime) {
				WL.LoadFromFile(wlfile)
				WLModTime = s.ModTime()
				goboot.Log.Infof("whitelist %v reload.", wlfile)
			}
			time.Sleep(10 * time.Second)
		}
	}()
	return nil
}

// 初始化函数
func init() {
	flag.StringVar(&RunEnv, "env", "dev", "app run env: [dev|dev-prod|prod]")
	flag.IntVar(&ListenPort, "port", 9000, "http listen port: [9000|9001]")
	flag.Parse()

	goboot.Init(RunEnv)
	goboot.OnAppStart(initWhitelist, 10)
	goboot.OnAppStart(initCache, 20)
	goboot.Startup()
}

func main() {
	if false {
		return
	}
	goboot.Log.Infof("run mode: %v\n", RunEnv)

	c := &Controller{}

	// n := negroni.Classic()
	n := negroni.New()

	n.Use(negroni.HandlerFunc(headerMiddleware))
	n.Use(negroni.HandlerFunc(crossRequestMiddleware))

	r := mux.NewRouter()
	r.KeepContext = false
	r.HandleFunc("/{rcType:(?:js|css)}", c.MergeHandler).Methods("GET", "POST", "HEAD")
	r.HandleFunc("/update", c.UpdateHandler).Methods("GET", "POST")

	n.UseHandler(r)

	srv := &http.Server{
		Handler: n,
		Addr:    fmt.Sprintf("0.0.0.0:%d", ListenPort),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		log.Fatal(srv.ListenAndServe())
	}()

	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)
	<-stopChan // wait for SIGINT
	log.Println("Shutting down server...")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	log.Fatal(srv.Shutdown(ctx))
	log.Println("Server gracefully stopped")

}
