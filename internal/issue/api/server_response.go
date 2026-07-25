package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

func writeStoreError(w http.ResponseWriter, err error, action string, resource string) {
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, action+".not_found", errors.New(resource+" not found"))
		return
	}
	writeError(w, http.StatusBadRequest, action+".invalid_input", err)
}

type responseMeta map[string]any

type successResponse struct {
	Data any          `json:"data"`
	Meta responseMeta `json:"meta"`
}

type errorResponse struct {
	Error responseError `json:"error"`
	Meta  responseMeta  `json:"meta"`
}

type responseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	writeRawJSON(w, status, successResponse{Data: value, Meta: responseMeta{}})
}

func writeJSONWithMeta(w http.ResponseWriter, status int, value any, meta responseMeta) {
	writeRawJSON(w, status, successResponse{Data: value, Meta: meta})
}

func writeError(w http.ResponseWriter, status int, code string, err error) {
	writeRawJSON(w, status, errorResponse{
		Error: responseError{Code: code, Message: err.Error()},
		Meta:  responseMeta{},
	})
}

func writeRawJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write json response: %v", err)
	}
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && (strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, PUT, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
