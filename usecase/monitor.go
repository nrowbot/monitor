package usecase

import (
	"context"
	"log"
	"sync"
	"time"

	"monitor/adapter/ping"
	"monitor/domain"
)

type MonitorUsecase struct {
	hostRepo map[string]*domain.Host
	pinger   ping.Pinger
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	verbose  bool
}

func NewMonitorUsecase(hostRepo map[string]*domain.Host, pinger ping.Pinger, interval time.Duration, verbose bool) *MonitorUsecase {
	ctx, cancel := context.WithCancel(context.Background())
	return &MonitorUsecase{
		hostRepo: hostRepo,
		pinger:   pinger,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
		verbose:  verbose,
	}
}

func (u *MonitorUsecase) Start() {
	for host := range u.hostRepo {
		u.wg.Add(1)
		go u.monitorHost(host)
	}
}

func (u *MonitorUsecase) Stop() {
	u.cancel()
	u.wg.Wait()
}

func (u *MonitorUsecase) monitorHost(host string) {
	defer u.wg.Done()
	ticker := time.NewTicker(u.interval)
	defer ticker.Stop()

	for {
		select {
		case <-u.ctx.Done():
			if u.verbose {
				log.Printf("[monitor] Stopped monitoring %s", host)
			}
			return
		case <-ticker.C:
			start := time.Now()
			latency, packetLoss, err := u.pinger.Ping(host)
			status := "up"
			switch {
			case err != nil || packetLoss >= 80:
				status = "down"
			case packetLoss >= 20:
				status = "degraded"
			}

			if err != nil || packetLoss == 100 {
				if u.verbose {
					if err != nil {
						log.Printf("[monitor] Error pinging %s: %v", host, err)
					} else {
						log.Printf("[monitor] No response from %s (100%% loss)", host)
					}
				}
				latency = 0
			}

			metrics := domain.Metrics{
				Latency:    latency,
				PacketLoss: packetLoss,
				Status:     status,
				Timestamp:  time.Now().Unix(),
			}
			u.hostRepo[host].AddMetrics(metrics)

			if u.verbose {
				elapsed := time.Since(start)
				log.Printf("[monitor] %s -> %s | latency: %.2fms | loss: %.2f%% | cycle took: %v",
					host, status, latency, packetLoss, elapsed)
			}
		}
	}
}
