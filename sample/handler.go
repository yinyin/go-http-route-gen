package main

import (
	"fmt"
	"io"
	"net/http"
)

type sampleHandler struct {
}

func (h *sampleHandler) responseText(w http.ResponseWriter, req *http.Request, message string) {
	header := w.Header()
	header.Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	txt := message + ": [" + req.RequestURI + "]\n"
	io.WriteString(w, txt)

}

func (h *sampleHandler) uniqueJSON(w http.ResponseWriter, req *http.Request, index, bound int, num int32) {
	h.responseText(w, req, fmt.Sprintf("uniqueJSON(num=%d)", num))
}

func (h *sampleHandler) uniqueText(w http.ResponseWriter, req *http.Request, index, bound int, num int32) {
	h.responseText(w, req, fmt.Sprintf("uniqueText(num=%d)", num))
}

func (h *sampleHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if routedIdent, err := h.routeRequest(w, req); nil != err {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	} else if routedIdent == RouteIncomplete {
		http.Error(w, "incomplete request URI", http.StatusBadRequest)
		return
	} else if routedIdent > RouteSuccess {
		return
	}
	h.responseText(w, req, "Last route")
}
