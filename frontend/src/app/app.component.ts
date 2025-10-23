import { Component, AfterViewInit, OnDestroy } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { interval, Subscription } from 'rxjs';

import { Chart } from 'chart.js';

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
  standalone: true, // Ensure this is present
  imports: [CommonModule], // Added for *ngFor and | keyvalue
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements AfterViewInit, OnDestroy {
  metrics: { [key: string]: HostMetrics } = {};
  private subscription: Subscription | undefined;
  private charts: { [key: string]: Chart } = {};

  constructor(private http: HttpClient) {}

  ngAfterViewInit(): void {
    this.updateDashboard();
    this.subscription = interval(5000).subscribe(() => this.updateDashboard());
  }

  ngOnDestroy(): void {
    if (this.subscription) {
      this.subscription.unsubscribe();
    }
    for (const key in this.charts) {
      this.charts[key].destroy();
    }
  }

  updateDashboard(): void {
    this.http.get<{ [key: string]: HostMetrics }>('/metrics').subscribe(
      data => {
        this.metrics = data;
        for (const host in data) {
          const chartId = `chart-${host}`;
          const ctx = document.getElementById(chartId) as HTMLCanvasElement;
          if (ctx) {
            if (this.charts[host]) {
              this.charts[host].destroy();
            }
            this.charts[host] = new Chart(ctx, {
              type: 'line',
              data: {
                labels: data[host].History.map(h => new Date(h.Timestamp * 1000).toLocaleTimeString()),
                datasets: [{
                  label: 'Latency (ms)',
                  data: data[host].History.map(h => h.Latency),
                  borderColor: 'blue',
                  fill: false
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
      },
      error => console.error('Error:', error)
    );
  }

  getStatusClass(status: string): string {
    return `status-${status}`;
  }
}