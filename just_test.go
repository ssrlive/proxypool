package main

import (
	"testing"

	"github.com/asdlokj1qpi23/proxypool/pkg/healthcheck"
)

func TestHello(t *testing.T) {
	got := "Hello, world"
	want := "Hello, world"

	if got != want {
		t.Errorf("got '%q' want '%q'", got, want)
	}

	healthcheck.PingFromChina("sina.com.cn", "443")
}
