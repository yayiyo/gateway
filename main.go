package main

import (
	"os"
	"os/signal"
	"syscall"

	"gateway/router"
	"gateway/utils"
)

func main() {
	utils.InitModule("./conf/dev/", []string{"base", "mysql", "redis"})
	defer utils.Destroy()
	router.HttpServerRun()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	router.HttpServerStop()
}
