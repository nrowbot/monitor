# README.md (updated Go version)

## Monitor

This is a Golang-based host monitoring tool that pings specified hosts at regular intervals, tracks metrics like latency and packet loss, and displays them on an Angular-based web dashboard.

### Setup Instructions

1. Install Go 1.25+: Download from https://go.dev/dl/.
2. Install Node.js 18+: Download from https://nodejs.org/.
3. Install Angular CLI: `npm install -g @angular/cli`.
4. Create Angular frontend: `mkdir frontend; cd frontend; ng new . --skip-git --routing=false --style=css` (choose defaults if prompted).
5. Apply the provided changes to Angular files (index.html, app.module.ts, etc.).
6. Run `go mod tidy` in root to fetch dependencies.
7. Build frontend: `make frontend-build`.
8. Build Go: `make build`.
9. Grant capabilities: `sudo setcap cap_net_raw+eip ./monitor`.
10. Run: `./monitor --hosts=google.com,example.com --port=8081 --interval=5`.
11. Access dashboard: http://localhost:8081.

For development: After changing Angular, run `make frontend-build` then restart the server.

### Docker

Build: `docker build -t monitor -f dockerfile .`

Run: `docker run -d -p 8081:8081 --cap-add=NET_RAW monitor --hosts=google.com --interval=5`

### Design Choices

- **Golang**: Chosen for concurrency (goroutines for per-host monitoring), performance, and static binaries.
- **Clean Architecture**: Layers separate domain logic from adapters (ping/web), enabling easy swapping (e.g., DB instead of in-memory).
- **Ping**: Uses go-ping library for ICMP. Requires NET_RAW cap for non-root.
- **Frontend with Angular**: Per user request, switched from vanilla HTML/JS to Angular for a more structured, component-based UI. This allows easier extension (e.g., adding routes, forms) if the dashboard grows. However, for this simple polling dashboard, vanilla JS would suffice and is lighter—Angular adds build steps and complexity but improves maintainability for larger apps. Separate 'frontend/' directory keeps backend and frontend concerns isolated.
- **Web Serving**: Go embeds Angular's built dist files for a single binary. Uses a simple SPA handler to serve index.html for root and fallback.
- **Polling**: Frontend polls /metrics every 5s via Angular's HttpClient and setInterval. Charts use Chart.js (CDN) directly in component.
- **Metrics**: In-memory history (last 10 pings) for averages. Expandable to persistent storage.
- **Systemd**: Daemonizes with restart.
- **Best Practices**: Flags for config, graceful shutdown, logging, tests. Multi-stage Docker for building frontend and backend.

Why Angular over embedded static? Provides better organization for UI logic, TypeScript for type safety, and easier scaling. Decisions prioritize separation for clean arch.

Tests: Basic, cover usecase/ping. Expand as needed.