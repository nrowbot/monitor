package domain

type Metrics struct {
	Latency    float64 // ms
	PacketLoss float64 // %
	Status     string  // "up" or "down"
	Timestamp  int64   // unix timestamp
}

type Host struct {
	Name           string
	MetricsHistory []Metrics // last N metrics
}

func (h *Host) LatestMetrics() Metrics {
	if len(h.MetricsHistory) == 0 {
		return Metrics{Status: "unknown"}
	}
	return h.MetricsHistory[len(h.MetricsHistory)-1]
}

func (h *Host) AvgLatency() float64 {
	if len(h.MetricsHistory) == 0 {
		return 0
	}
	sum := 0.0
	for _, m := range h.MetricsHistory {
		sum += m.Latency
	}
	return sum / float64(len(h.MetricsHistory))
}

func (h *Host) AvgPacketLoss() float64 {
	if len(h.MetricsHistory) == 0 {
		return 0
	}
	sum := 0.0
	for _, m := range h.MetricsHistory {
		sum += m.PacketLoss
	}
	return sum / float64(len(h.MetricsHistory))
}

func (h *Host) AddMetrics(m Metrics) {
	h.MetricsHistory = append(h.MetricsHistory, m)
	if len(h.MetricsHistory) > 10 { // Keep last 10 for metrics calculation
		h.MetricsHistory = h.MetricsHistory[1:]
	}
}
