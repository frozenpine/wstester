package server

import (
	"testing"
)

func TestGetNS(t *testing.T) {
	cfg := NewSvrConfig()

	t.Log(cfg.GetNS())

	cfg.FrontID = "1"
	t.Log(cfg.GetNS())

	cfg.ChangeListen("1.1.1.1:9988")

	t.Log(cfg.GetNS())
}
