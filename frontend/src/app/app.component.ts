import { Component, AfterViewInit, OnDestroy, ChangeDetectorRef } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { interval, Subscription } from 'rxjs';
import { Chart } from 'chart.js/auto';

interface Metrics {
  Latency: number;
  PacketLoss: number;
  Status: string;
  Timestamp: number;
}

interface HostMetrics {
  Status: string;
  AvgLatency: number;
  AvgPacketLoss: number;
  History: Metrics[];
}

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements AfterViewInit, OnDestroy {
  metrics: { [key: string]: HostMetrics } = {};
  private subscription?: Subscription;
  private charts: { [key: string]: Chart } = {};

  constructor(private http: HttpClient, private cdr: ChangeDetectorRef) {}

  ngAfterViewInit(): void {
    this.updateDashboard();
    this.subscription = interval(5000).subscribe(() => this.updateDashboard());
  }

  ngOnDestroy(): void {
    this.subscription?.unsubscribe();
    for (const chart of Object.values(this.charts)) {
      chart.destroy();
    }
  }

  updateDashboard(): void {
    this.http.get<{ [key: string]: HostMetrics }>('/metrics').subscribe({
      next: data => {
        this.metrics = data;
        this.cdr.detectChanges();
        setTimeout(() => {
          this.updateCharts(data);
          this.updateDashboardChart();
        }, 0);
      },
      error: err => console.error('Error fetching metrics:', err)
    });
  }

  private updateCharts(data: { [key: string]: HostMetrics }): void {
    for (const host in data) {
      const safeHostId = this.sanitizeId(host);
      const canvas = document.getElementById(`chart-${safeHostId}`) as HTMLCanvasElement | null;
      if (!canvas) continue;
  
      const history = data[host].History;
      const labels = history.map(h => new Date(h.Timestamp * 1000).toLocaleTimeString());
      const values = history.map(h => h.Latency);
  
      if (this.charts[host]) {
        const chart = this.charts[host];
        chart.data.labels = labels;
        chart.data.datasets[0].data = values;
        chart.update('none');
      } else {
        const context = canvas.getContext('2d');
        if (!context) continue;
  
        this.charts[host] = new Chart(context, {
          type: 'line',
          data: {
            labels,
            datasets: [{
              label: 'Latency (ms)',
              data: values,
              borderColor: 'blue',
              fill: false,
              tension: 0.2
            }]
          },
          options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
              y: { beginAtZero: true }
            }
          }
        });
      }
    }
  }

  private dashboardChart?: Chart;
  private avgLatencyHistory: number[] = [];
  private timeLabels: string[] = [];

  private updateDashboardChart(): void {
    const summary = this.getDashboardSummary();
    const timestamp = new Date().toLocaleTimeString();
    this.avgLatencyHistory.push(summary.avgLatency);
    this.timeLabels.push(timestamp);

    if (this.avgLatencyHistory.length > 20) {
      this.avgLatencyHistory.shift();
      this.timeLabels.shift();
    }

    const canvas = document.getElementById('dashboard-chart') as HTMLCanvasElement;
    if (!canvas) return;

    if (this.dashboardChart) {
      this.dashboardChart.data.labels = this.timeLabels;
      this.dashboardChart.data.datasets[0].data = this.avgLatencyHistory;
      this.dashboardChart.update('none');
    } else {
      const ctx = canvas.getContext('2d');
      if (!ctx) return;
      this.dashboardChart = new Chart(ctx, {
        type: 'line',
        data: {
          labels: this.timeLabels,
          datasets: [{
            label: 'Average Latency (ms)',
            data: this.avgLatencyHistory,
            borderColor: 'green',
            fill: false,
            tension: 0.3
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          scales: {
            y: { beginAtZero: true }
          },
          plugins: { legend: { display: false } }
        }
      });
    }
  }


  getDashboardSummary() {
    const hosts = Object.keys(this.metrics);
    if (hosts.length === 0) {
      return {
        total: 0,
        up: 0,
        down: 0,
        avgLatency: 0,
        avgPacketLoss: 0,
        worstHost: ''
      };
    }

    let up = 0;
    let down = 0;
    let totalLatency = 0;
    let totalLoss = 0;
    let countLatency = 0;

    let worstHost = '';
    let worstLatency = -1;

    for (const host of hosts) {
      const m = this.metrics[host];
      if (m.Status === 'up') up++;
      else down++;

      if (m.AvgLatency > 0) {
        totalLatency += m.AvgLatency;
        countLatency++;
      }
      totalLoss += m.AvgPacketLoss;

      if (m.AvgLatency > worstLatency) {
        worstLatency = m.AvgLatency;
        worstHost = host;
      }
    }

    const uptimePct = (up / hosts.length) * 100;
    let health = 'Good';
    if (uptimePct < 90 || totalLoss / hosts.length > 10) health = 'Degraded';
    if (uptimePct < 50) health = 'Critical';

    return {
      total: hosts.length,
      up,
      down,
      avgLatency: countLatency > 0 ? totalLatency / countLatency : 0,
      avgPacketLoss: totalLoss / hosts.length,
      worstHost,
      health, 
      uptimePct
    };
  }

  getStatusClass(status: string): string {
    return `status-${status}`;
  }

  sanitizeId(host: string): string {
    return host.replace(/[^a-zA-Z0-9_-]/g, '');
  }  

  trackByHost(index: number, item: { key: string }): string {
    return item.key;
  }  
}
