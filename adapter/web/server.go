package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path"
	"sync"

	"monitor/domain"
)

type Server struct {
	port     int
	hostRepo map[string]*domain.Host
	mutex    sync.RWMutex
	server   *http.Server
	content  fs.FS
}

func NewServer(port int, hostRepo map[string]*domain.Host, distFS embed.FS) *Server {
	content, err := fs.Sub(distFS, "frontend/dist/frontend/browser")
	entries, _ := fs.ReadDir(content, ".")
	for _, e := range entries {
		log.Println("Embedded file:", e.Name())
	}

	if err != nil {
		log.Fatal("Failed to sub embed FS: ", err)
	}
	return &Server{
		port:     port,
		hostRepo: hostRepo,
		content:  content,
	}
}

func (s *Server) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", s.serveMetrics)
	mux.HandleFunc("/", s.serveSPA)

	s.server = &http.Server{Addr: fmt.Sprintf(":%d", s.port), Handler: mux}
	log.Printf("Starting web server on port %d", s.port)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Web server error: %v", err)
	}
}

func (s *Server) Stop() {
	if err := s.server.Close(); err != nil {
		log.Printf("Error stopping web server: %v", err)
	}
}

func (s *Server) serveSPA(w http.ResponseWriter, r *http.Request) {
	p := path.Clean(r.URL.Path)
	if p == "/" || p == "" {
		p = "index.html"
	} else {
		p = p[1:] // remove leading /
	}

	data, err := fs.ReadFile(s.content, p)
	if err != nil {
		// Fallback to index.html for SPA routing
		data, err = fs.ReadFile(s.content, "index.html")
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
		return
	}

	// Serve the file with appropriate content type
	switch path.Ext(p) {
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Write(data)
}

func (s *Server) serveMetrics(w http.ResponseWriter, r *http.Request) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	metricsMap := make(map[string]struct {
		Status        string
		AvgLatency    float64
		AvgPacketLoss float64
		History       []domain.Metrics
	})
	for name, host := range s.hostRepo {
		metricsMap[name] = struct {
			Status        string
			AvgLatency    float64
			AvgPacketLoss float64
			History       []domain.Metrics
		}{
			Status:        host.LatestMetrics().Status,
			AvgLatency:    host.AvgLatency(),
			AvgPacketLoss: host.AvgPacketLoss(),
			History:       host.MetricsHistory,
		}
	}

	jsonData, err := json.Marshal(metricsMap)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
