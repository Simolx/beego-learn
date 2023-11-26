package main

import (
	"fmt"
	"net"
	"github.com/beego/beego/v2/server/web"
	"example.com/dev/proxy"
	"os"
	"time"
)

type UserController struct {
	web.Controller
}

func (u *UserController) HelloWorld() {
	u.Ctx.WriteString(fmt.Sprint("hello, $s", u.Ctx.Request.RemoteAddr))
}

func main() {
	ipAddr := os.Args[1]
	port := os.Args[2]
	// now you start the beego as http server.
	// it will listen to port 8080
	web.AutoRouter(&UserController{})
//	web.Run(fmt.Sprintf("%s:%s", ipAddr, port))

	// it will listen to 8080
	// beego.Run("localhost")

	// it will listen to 8089
	// beego.Run(":8089")

	// it will listen to 8089
	// beego.Run("127.0.0.1:8089")
	address := fmt.Sprintf("%s:%s", ipAddr, port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	proxyListener := &proxy.Listener{
		Listener:          ln,
		ProxyHeaderTimeout: 10 * time.Second,
	}
	defer proxyListener.Close()

		
	web.BeeApp.Server.Serve(proxyListener)
	web.BeeApp.Run("localhost:1234")
}