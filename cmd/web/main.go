package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

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

	log.Printf("web server listening on %s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("web server failed: %v", err)
	}
}

func newMux(trackerURL string, orchestratorURL string) (http.Handler, error) {
	dist, err := fs.Sub(embeddedAssets, "frontend/dist")
	if err != nil {
		return nil, fmt.Errorf("open embedded frontend assets: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/tracker/", reverseProxy("/tracker", trackerURL))
	mux.Handle("/orchestrator/", reverseProxy("/orchestrator", orchestratorURL))
	mux.Handle("/", spaHandler{assets: dist})
	return mux, nil
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
