package main

import (
	"os"
	"fmt"
	"net"
	"net/http"
	"example.com/dev/proxy"
	"time"
)

func main() {
	ipAddr := os.Args[1]
	port := os.Args[2]
	server := http.Server{
		Addr: fmt.Sprintf("%s:%s", ipAddr, port),
		}

		ln, err := net.Listen("tcp", server.Addr)
		if err != nil {
			panic(err)
		}


		http.HandleFunc("/devbar", func(w http.ResponseWriter, r *http.Request) {
			ipSource := r.RemoteAddr
			fmt.Fprintf(w, "Hello, %s", ipSource)
		})

		proxyListener := &proxy.Listener{
			Listener:          ln,
			ProxyHeaderTimeout: 10 * time.Second,
			}
			defer proxyListener.Close()

		server.Serve(proxyListener)
}

