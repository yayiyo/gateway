package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gateway/dao"
	"gateway/grpcproxy"
	"gateway/httpproxy"
	"gateway/router"
	"gateway/tcpproxy"
	"gateway/utils"
)

//endpoint dashboard后台管理  server代理服务器
//config ./conf/prod/ 对应配置文件夹

var (
	endpoint = flag.String("endpoint", "", "input endpoint dashboard or server")
	config   = flag.String("config", "", "input config file like ./conf/dev/")
)

// @host localhost:8080
// @BasePath /api/v1
// @query.collection.format multi

// @securityDefinitions.basic BasicAuth

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	flag.Parse()
	if *endpoint == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *config == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *endpoint == "dashboard" {
		err := utils.InitModule(*config)
		if err != nil {
			log.Fatal(err)
		}
		defer utils.Destroy()
		router.HttpServerRun()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		router.HttpServerStop()
	} else {
		err := utils.InitModule(*config)
		if err != nil {
			log.Fatal(err)
		}
		defer utils.Destroy()
		dao.ServiceManagerHandler.LoadOnce()
		dao.AppManagerHandler.LoadOnce()

		go func() {
			httpproxy.HttpServerRun()
		}()
		go func() {
			httpproxy.HttpsServerRun()
		}()
		go func() {
			tcpproxy.TcpServerRun()
		}()
		go func() {
			grpcproxy.GrpcServerRun()
		}()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		tcpproxy.TcpServerStop()
		grpcproxy.GrpcServerStop()
		httpproxy.HttpServerStop()
		httpproxy.HttpsServerStop()
	}
}
