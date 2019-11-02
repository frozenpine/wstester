package main

import (
	"context"

	"github.com/frozenpine/wstester/server"
	"github.com/frozenpine/wstester/utils/log"
	"go.uber.org/zap/zapcore"
)

func main() {
	log.SetLogLevel(zapcore.DebugLevel)

	cfg := server.NewConfig()

	svr := server.NewServer(nil, cfg)

	svr.RunForever(context.Background())
}
