package main

import (
	"github.com/vultisig/vultisigner/api"
	"github.com/vultisig/vultisigner/config"
)

func main() {
	port := config.AppConfig.Server.Port
	server := api.NewServer(port)
	if err := server.StartServer(); err != nil {
		panic(err)
	}
}
