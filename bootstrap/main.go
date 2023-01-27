package main

import (
	"example.com/lx/beego/dev/utils"
	"fmt"
	"os"
	"os/exec"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

var logger = logs.NewLogger()

type ServerController struct {
	web.Controller
}

func (c ServerController) HealthCheck() {
	c.Ctx.Output.Body([]byte(`{"result": "start server succeed"}`))
}

func runCommand(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	resp, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("start zk Server failed, error %v", err)
	}
	return resp, err
}

func (c ServerController) StartServer() {
	if resp, err := runCommand("/opt/zookeeper/bin/zkServer.sh", "start"); err != nil {
		logger.Error("start zk Server failed, error %v", err)
		c.Ctx.Output.Body(resp)
		return
	}
	c.Ctx.Output.Body([]byte(`{"result": "start zk succeed"}`))
}

func (c ServerController) StopServer() {
	if resp, err := runCommand("/opt/zookeeper/bin/zkServer.sh", "stop"); err != nil {
		logger.Error("stop zk Server failed, error %v", err)
		c.Ctx.Output.Body(resp)
		return
	}
	c.Ctx.Output.Body([]byte(`{"result": "stop zk succeed"}`))
}

func main() {
	logs.Info("run server args, %v", os.Args)
	if len(os.Args) == 1 {
		logs.Error("need argunents")
		return
		//    } else if os.Args[1] == "cert" {
		//        ipv4String := "192.168.0.104"
		//        serviceName := "KafkaService"
		//        GenerateCert(ipv4String, serviceName)
		//        return
	} else if os.Args[1] == "https" {
		web.BConfig.Listen.EnableHTTP = false
		if len(os.Args) == 3 {
			certConfig := utils.CertConfig{
				ServerCert: fmt.Sprintf("conf/cert%s/server%s.crt", os.Args[2], os.Args[2]),
				ServerKey:  fmt.Sprintf("conf/cert%s/server%s.key", os.Args[2], os.Args[2]),
				CaCert:     fmt.Sprintf("conf/cert%s/ca.crt", os.Args[2]),
			}
			web.BConfig.Listen.HTTPSCertFile = certConfig.ServerCert
			web.BConfig.Listen.HTTPSKeyFile = certConfig.ServerKey
			web.BConfig.Listen.TrustCaFile = certConfig.CaCert
			web.BConfig.Listen.HTTPSAddr = "192.168.0.104"
		}
		logger.Info("start server")
		web.CtrlGet("/server/health", ServerController.HealthCheck)
		web.CtrlPost("/server/start", ServerController.StartServer)
		web.CtrlPost("/server/stop", ServerController.StopServer)
		logger.Info("server handlers %v", web.PrintTree())
		web.Run()
	} else if os.Args[1] == "client" {
		url := "https://KafkaService:8010/server/health"
		certConfig := &utils.CertConfig{
			ServerCert: fmt.Sprintf("conf/cert%s/server%s.crt", os.Args[2], os.Args[2]),
			ServerKey:  fmt.Sprintf("conf/cert%s/server%s.key", os.Args[2], os.Args[2]),
			CaCert:     fmt.Sprintf("conf/cert%s/ca.crt", os.Args[2]),
		}
		utils.GetRequest(url, certConfig)
	} else if os.Args[1] == "httpsdev" {
		web.BConfig.Listen.EnableHTTP = false
		//        if len(os.Args) == 3 {
		certConfig := utils.CertConfig{
			ServerCert: fmt.Sprintf("conf/certs/server%s.crt", os.Args[2]),
			ServerKey:  fmt.Sprintf("conf/certs/server%s.key", os.Args[2]),
			CaCert:     fmt.Sprintf("conf/certs/ca%s.crt", os.Args[2]),
		}
		web.BConfig.Listen.HTTPSCertFile = certConfig.ServerCert
		web.BConfig.Listen.HTTPSKeyFile = certConfig.ServerKey
		web.BConfig.Listen.TrustCaFile = certConfig.CaCert
		web.BConfig.Listen.HTTPSAddr = "127.0.0.1"
		//        }
		logger.Info("start server")
		web.CtrlGet("/server/health", ServerController.HealthCheck)
		web.CtrlPost("/server/start", ServerController.StartServer)
		web.CtrlPost("/server/stop", ServerController.StopServer)
		logger.Info("server handlers %v", web.PrintTree())
		web.Run()
	} else if os.Args[1] == "clientdev" {
		url := "https://127.0.0.1:8010/server/health"
		certConfig := &utils.CertConfig{
			ServerCert: fmt.Sprintf("conf/certs/client%s.crt", os.Args[2]),
			ServerKey:  fmt.Sprintf("conf/certs/client%s.key", os.Args[2]),
			CaCert:     fmt.Sprintf("conf/certs/ca%s.crt", os.Args[2]),
		}
		utils.GetRequest(url, certConfig)
	}
}
