package server

import (
	"testing"
)

func TestGetNS(t *testing.T) {
	cfg := NewConfig()

	t.Log(cfg.GetNS())

	cfg.FrontID = "1"
	t.Log(cfg.GetNS())

	cfg.ChangeListen("1.1.1.1:9988")

	t.Log(cfg.GetNS())
}
