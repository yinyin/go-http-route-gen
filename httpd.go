package main

import (
	"net/http"
	"strconv"
)

type httpHandler struct {
	JSONContent       []byte
	ContentLengthText string
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", h.ContentLengthText)
	w.WriteHeader(http.StatusOK)
	w.Write(h.JSONContent)
}

func runHTTPService(httpAddr string, jsonContent []byte) error {
	h := httpHandler{
		JSONContent:       jsonContent,
		ContentLengthText: strconv.FormatInt(int64(len(jsonContent)), 10),
	}
	s := &http.Server{
		Addr:    httpAddr,
		Handler: &h,
	}
	return s.ListenAndServe()
}
