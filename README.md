# README.md (updated Go version)

## Monitor

This is a Golang-based host monitoring tool that pings specified hosts at regular intervals, tracks metrics like latency and packet loss, and displays them on an Angular-based web dashboard.

### Setup Instructions

1. Install Go: Download from https://go.dev/dl/.
2. Install Node.js: Download from https://nodejs.org/.
3. Install Angular CLI: `npm install -g @angular/cli`.
4. Run `go mod tidy` in root to fetch dependencies.
5. Build frontend: `make frontend-build`.
6. Build Go: `make build`.
7. Grant capabilities: `sudo setcap cap_net_raw+eip ./monitor`.
8. Run: `./monitor --hosts=google.com,example.com --port=8081 --interval=5`.
9. Access dashboard: http://localhost:8081.

For development: After changing Angular, run `make frontend-build` then restart the server.

### Docker
Note: This would need additional troubleshooting to work on all systems. Currently on MacOS, for example, it will only scan Internet hosts, not local network hosts due to the native Docker networking set up on MacOS.

Build: `docker build -t monitor -f Dockerfile .`

Run: `docker run -d -p 8081:8081 --cap-add=NET_RAW monitor --hosts=google.com --interval=5`

### Design Choices

- **Golang**: Chosen for concurrency (goroutines for per-host monitoring), performance, and static binaries.
- **Clean Architecture**: Layers separate domain logic from adapters (ping/web), enabling easy swapping (e.g., DB instead of in-memory).
- **Ping**: Uses pro-bing library for ICMP. Requires NET_RAW cap for non-root.
- **Frontend with Angular**: Chosen for a more structured, component-based UI. This allows easier extension (e.g., adding routes, forms) if the dashboard grows. However, for this simple polling dashboard, vanilla JS would suffice and is lighter—Angular adds build steps and complexity but improves maintainability for larger apps. Separate 'frontend/' directory keeps backend and frontend concerns isolated.
- **Web Serving**: Go embeds Angular's built dist files for a single binary. Uses a simple SPA handler to serve index.html for root and fallback.
- **Polling**: Frontend polls /metrics every 5s via Angular's HttpClient and setInterval.
- **Metrics**: In-memory history (last 10 pings) for averages. Expandable to persistent storage.
- **Systemd**: Daemonizes with restart.
- **Best Practices**: Flags for config, graceful shutdown, logging (with --verbose). Multi-stage Docker for building frontend and backend.

Tests: Basic, covers frontend. Expand as needed