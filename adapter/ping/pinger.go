package ping

import (
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

type Pinger interface {
	Ping(host string) (latency float64, packetLoss float64, err error)
}

type ICMPPinger struct{}

func NewPinger() Pinger {
	return &ICMPPinger{}
}

func (p *ICMPPinger) Ping(host string) (float64, float64, error) {
	pinger, err := probing.NewPinger(host)
	if err != nil {
		return 0, 0, err
	}
	pinger.Count = 3 // Send 3 packets
	pinger.Timeout = time.Second * 5
	err = pinger.Run()
	if err != nil {
		return 0, 0, err
	}
	stats := pinger.Statistics()
	if stats.PacketsRecv == 0 {
		return 0, 100, nil // All lost
	}
	return float64(stats.AvgRtt.Milliseconds()), stats.PacketLoss, nil
}
