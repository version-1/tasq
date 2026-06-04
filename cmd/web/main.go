package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/version-1/tasq/internal/config"
)

//go:embed frontend/dist
var embeddedAssets embed.FS

func main() {
	addr := flag.String("addr", fmt.Sprintf(":%d", config.DefaultWebPort), "web server listen address")
	trackerURL := flag.String("tracker-url", fmt.Sprintf("http://127.0.0.1:%d", config.DefaultIssueTrackerPort), "issue-tracker base URL")
	orchestratorURL := flag.String("orchestrator-url", fmt.Sprintf("http://127.0.0.1:%d", config.DefaultOrchestratorPort), "orchestrator base URL")
	flag.Parse()

	mux, err := newMux(*trackerURL, *orchestratorURL)
	if err != nil {
		log.Fatalf("initialize web server: %v", err)
	}

	if _, err := config.EnsureHome(); err != nil {
		log.Fatalf("ensure TQ_HOME: %v", err)
	}
	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	serviceAddr := clientAddr(listener.Addr().String())
	if err := config.UpdateState(func(state *config.State) error {
		state.Web = &config.ServiceState{
			PID:       os.Getpid(),
			Addr:      serviceAddr,
			StartedAt: time.Now().UTC(),
		}
		return nil
	}); err != nil {
		log.Fatalf("write web state: %v", err)
	}
	defer func() {
		if err := config.UpdateState(func(state *config.State) error {
			state.Web = nil
			return nil
		}); err != nil {
			log.Printf("clear web state: %v", err)
		}
	}()

	go func() {
		log.Printf("web server listening on %s", serviceAddr)
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("web server failed: %v", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown server: %v", err)
	}
}

func newMux(trackerURL string, orchestratorURL string) (http.Handler, error) {
	dist, err := fs.Sub(embeddedAssets, "frontend/dist")
	if err != nil {
		return nil, fmt.Errorf("open embedded frontend assets: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/tracker/", reverseProxy("/tracker", trackerURL))
	mux.Handle("/orchestrator/", reverseProxy("/orchestrator", orchestratorURL))
	mux.Handle("/", spaHandler{assets: dist})
	return mux, nil
}

func clientAddr(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	if host == "" || host == "::" || host == "[::]" || host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, port)
}

func reverseProxy(prefix string, target string) http.Handler {
	targetURL, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}
		req.Host = targetURL.Host
	}
	return proxy
}

type spaHandler struct {
	assets fs.FS
}

func (handler spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	info, err := fs.Stat(handler.assets, path)
	if err == nil && !info.IsDir() {
		http.FileServer(http.FS(handler.assets)).ServeHTTP(w, r)
		return
	}

	index, err := fs.ReadFile(handler.assets, "index.html")
	if err != nil {
		http.Error(w, "frontend index is not available", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(index)
}
