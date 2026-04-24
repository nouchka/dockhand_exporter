package main

import (
	"net/http"
	"os"
)

func main() {
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":9090"
	}
	resp, err := http.Get("http://localhost" + addr + "/healthz")
	if err != nil || resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
	_ = resp.Body.Close()
}
