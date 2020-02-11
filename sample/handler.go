package main

import (
	"fmt"
	"io"
	"net/http"
)

type sampleHandler struct {
}

func (h *sampleHandler) responseText(w http.ResponseWriter, req *http.Request, pathOffset int, message string) {
	header := w.Header()
	header.Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	var restPath string
	if len(req.URL.Path) > pathOffset {
		restPath = string(req.URL.Path[pathOffset:])
	} else {
		restPath = fmt.Sprintf("-EMPTY; len(req.URL.Path)=%d; pathOffset=%d", len(req.URL.Path), pathOffset)
	}
	txt := message + ": [" + req.URL.Path + "]\nrest-path: [" + restPath + "]\n"
	io.WriteString(w, txt)
}

func (h *sampleHandler) queryProduct(w http.ResponseWriter, req *http.Request, pathOffset int, productName string) {
	h.responseText(w, req, pathOffset, fmt.Sprintf("queryProduct(productName=%s)", productName))
}

func (h *sampleHandler) downloadProduct(w http.ResponseWriter, req *http.Request, pathOffset int, sessionID, targetID int64) {
	h.responseText(w, req, pathOffset, fmt.Sprintf("downloadProduct(sessionId=%d, targetId=%d)", sessionID, targetID))
}

func (h *sampleHandler) listProducts(w http.ResponseWriter, req *http.Request, pathOffset int) {
	h.responseText(w, req, pathOffset, "listProducts()")
}

func (h *sampleHandler) showProduct(w http.ResponseWriter, req *http.Request, pathOffset int, productID int64) {
	h.responseText(w, req, pathOffset, fmt.Sprintf("showProduct(productId=%d)", productID))
}

func (h *sampleHandler) sampleData(w http.ResponseWriter, req *http.Request, pathOffset int) {
	h.responseText(w, req, pathOffset, "sampleData()")
}

func (h *sampleHandler) debugText(w http.ResponseWriter, req *http.Request, pathOffset int) {
	h.responseText(w, req, pathOffset, "debugText()")
}

func (h *sampleHandler) debugJSON(w http.ResponseWriter, req *http.Request, pathOffset int) {
	h.responseText(w, req, pathOffset, "debugJSON()")
}

func (h *sampleHandler) exactText(w http.ResponseWriter, req *http.Request, pathOffset int) {
	h.responseText(w, req, pathOffset, "exactText()")
}

func (h *sampleHandler) debugNumber(w http.ResponseWriter, req *http.Request, pathOffset int, num, hex1 int32, hex2 uint32) {
	h.responseText(w, req, pathOffset, fmt.Sprintf("debugNumber(num=%d, hex1=%v, hex2=%v)", num, hex1, hex2))
}

func (h *sampleHandler) uniqueText(w http.ResponseWriter, req *http.Request, pathOffset int, num int32) {
	h.responseText(w, req, pathOffset, fmt.Sprintf("uniqueText(num=%d)", num))
}

func (h *sampleHandler) uniqueJSON(w http.ResponseWriter, req *http.Request, pathOffset int, num int32) {
	h.responseText(w, req, pathOffset, fmt.Sprintf("uniqueJSON(num=%d)", num))
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
	} else if routedIdent == RouteMethodNotAllowed {
		return
	}
	h.responseText(w, req, 0, "Last route")
}
