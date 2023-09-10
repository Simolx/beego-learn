package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"example.com/lx/beego/dev/utils"

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

func (c ServerController) BenchMarkUnMarshal() {
	var requestBody map[string]string
	var data map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestBody); err != nil {
		logger.Error("Unmarshal request body failed, error: %v", err)
		c.Ctx.Output.Body([]byte(`{"result": "get request body failed"}`))
	} else if content, err := os.ReadFile(requestBody["path"]); err != nil {
		logger.Error("read file %s failed, err %v", requestBody["path"], err)
		c.Ctx.Output.Body([]byte(`{"result": "read file failed"}`))
	} else if err := json.Unmarshal(content, &data); err != nil {
		logger.Error("Unmarshal file content failed, error: %v", err)
		c.Ctx.Output.Body([]byte(`{"result": "unmarshal file content failed"}`))
	} else {
		logger.Info("there are %d items in file", len(data))
		c.Ctx.Output.Body([]byte(`{"result": "succeed"}`))
	}
}

func (c ServerController) BenchMarkTarCmd() {
	var requestBody map[string]string
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestBody); err != nil {
		logger.Error("Unmarshal request body failed, error: %v", err)
		c.Ctx.Output.Body([]byte(`{"result": "get request body failed"}`))
		return
	}
	if _, err := os.Stat(requestBody["path"]); err != nil {
		logger.Error("check file stat failed, error: %v", err)
		c.Ctx.Output.Body([]byte(`{"result": "file stat error"}`))
		return
	}
	tarParams, ok := requestBody["param"]
	if !ok {
		tarParams = "czvf"
	}
	cmd := exec.Command("tar", "-"+tarParams, "/tmp/result.tar.gz", "-C", requestBody["path"], ".")
	if _, err := cmd.CombinedOutput(); err != nil {
		logger.Error("check file stat failed, error: %v", err)
		c.Ctx.Output.Body([]byte(`{"result": "file stat error"}`))
		return
	}
	logger.Info("tar %s succeed", tarParams)
	c.Ctx.Output.Body([]byte(`{"result": "succeed"}`))
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
	} else if os.Args[1] == "http" {
		web.BConfig.Listen.EnableHTTPS = false
		web.CtrlPost("/check/unmarshal", ServerController.BenchMarkUnMarshal)
		web.CtrlPost("/check/compress", ServerController.BenchMarkTarCmd)
		go func() {
			logger.Info("start pprof result: %v", http.ListenAndServe("localhost:6060", nil))
		}()
		logger.Info("server handlers %v", web.PrintTree())
		web.Run()
	}
}
