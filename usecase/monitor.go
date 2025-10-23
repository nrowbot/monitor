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
}

func NewMonitorUsecase(hostRepo map[string]*domain.Host, pinger ping.Pinger, interval time.Duration) *MonitorUsecase {
	ctx, cancel := context.WithCancel(context.Background())
	return &MonitorUsecase{
		hostRepo: hostRepo,
		pinger:   pinger,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
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
			return
		case <-ticker.C:
			latency, packetLoss, err := u.pinger.Ping(host)
			status := "up"
			if err != nil {
				log.Printf("Error pinging %s: %v", host, err)
				status = "down"
				latency = 0
				packetLoss = 100
			}
			metrics := domain.Metrics{
				Latency:    latency,
				PacketLoss: packetLoss,
				Status:     status,
				Timestamp:  time.Now().Unix(),
			}
			u.hostRepo[host].AddMetrics(metrics)
		}
	}
}
