package main

import (
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

func HttpGet(url string) ([]byte, error) {
	var client = &http.Client{
		Timeout: time.Second * 15,
		Transport: &http.Transport{
			Dial:                (&net.Dialer{Timeout: 15 * time.Second}).Dial,
			TLSHandshakeTimeout: 15 * time.Second,
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
