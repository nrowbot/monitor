package main

import (
	"embed"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"monitor/adapter/ping"
	"monitor/adapter/web"
	"monitor/domain"
	"monitor/usecase"
)

//go:embed frontend/dist/frontend
var distFS embed.FS

func main() {
	hostsStr := flag.String("hosts", "", "Comma-separated list of hosts to monitor (required)")
	port := flag.Int("port", 8081, "Port for the web server")
	interval := flag.Int("interval", 5, "Ping interval in seconds")
	flag.Parse()

	if *hostsStr == "" {
		log.Fatal("Hosts flag is required")
	}

	hosts := strings.Split(*hostsStr, ",")
	hostRepo := make(map[string]*domain.Host)
	for _, h := range hosts {
		hostRepo[h] = &domain.Host{Name: h, MetricsHistory: make([]domain.Metrics, 0, 10)}
	}

	pinger := ping.NewPinger()
	monitorUsecase := usecase.NewMonitorUsecase(hostRepo, pinger, time.Duration(*interval)*time.Second)

	go monitorUsecase.Start()

	webServer := web.NewServer(*port, hostRepo, distFS)
	go webServer.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	monitorUsecase.Stop()
	webServer.Stop()
}
