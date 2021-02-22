package main

import (
	"github.com/netooo/go-tutorial-gopherjs/backend"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	u, _ := url.Parse("http://localhost:8080")
	rp := httputil.NewSingleHostReverseProxy(u)
	http.Handle("/", rp)
	http.Handle("/api/", backend.New())
	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("listen:", l.Addr())
	if err := http.Serve(l, nil); err != nil {
		log.Fatalln(err)
	}
}
