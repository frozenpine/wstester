package main

import (
	"context"

	"github.com/frozenpine/wstester/server"
)

func main() {
	cfg := server.NewConfig()

	svr := server.NewServer(nil, cfg)

	svr.RunForever(context.Background())
}
