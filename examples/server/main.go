package main

import (
	"context"

	"github.com/frozenpine/wstester/server"
	"github.com/frozenpine/wstester/utils/log"
)

func main() {
	log.SetLogLevel(log.DebugLevel)

	cfg := server.NewConfig()

	svr := server.NewServer(nil, cfg)

	svr.RunForever(context.Background())
}
