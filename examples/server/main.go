package main

import (
	"context"
	"log"

	"github.com/frozenpine/wstester/server"
)

func main() {
	log.SetFlags(log.Lmicroseconds | log.Ldate)

	cfg := server.NewConfig()

	svr := server.NewServer(nil, cfg)

	svr.RunForever(context.Background())
}
