package main

import (
	"fmt"
	"log"
	"lustre_auto/config"
	"lustre_auto/router"
	"net/http"
	"os"
)

func main() {
	endPoint := fmt.Sprintf(":%d", config.ConfigData.Server.Port)
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
