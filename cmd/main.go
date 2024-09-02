package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"lustre_auto/router"
	"net/http"
	"os"
)

func main() {
	gin.SetMode(setting.ServerSetting.RunMode)
	endPoint := fmt.Sprintf(":%d", setting.ServerSetting.HttpPort)

	server := &http.Server{
		Addr:    endPoint,
		Handler: router.InitRouter(),
		//ReadTimeout:    setting.ServerSetting.ReadTimeout,
		//WriteTimeout:   setting.ServerSetting.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("[info] start http server listening %s", endPoint)

	if err := server.ListenAndServe(); err != nil {
		os.Exit(1)
	}
}
