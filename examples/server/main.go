package main

import (
	"context"

	"github.com/frozenpine/wstester/server"
)

func main() {
	cfg := server.NewWsConfig()
	svr := server.NewServer(cfg)

	svr.RunForever(context.Background())
}
