package main

import (
	"log"
	"net/http"
	"os"

	"github.com/nouchka/dockhand_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	dockhandURL := os.Getenv("DOCKHAND_URL")
	if dockhandURL == "" {
		log.Fatal("DOCKHAND_URL environment variable is required")
	}

	dockhandToken := os.Getenv("DOCKHAND_TOKEN")
	if dockhandToken == "" {
		log.Fatal("DOCKHAND_TOKEN environment variable is required")
	}

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":9090"
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(collector.New(dockhandURL, dockhandToken))

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
<head><title>Dockhand Exporter</title></head>
<body>
<h1>Dockhand Exporter</h1>
<p><a href="/metrics">Metrics</a></p>
</body>
</html>`))
	})

	log.Printf("Starting dockhand_exporter on %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
