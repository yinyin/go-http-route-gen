package main

import (
	"errors"
	"net/http"
)

// go:generate stringer -type RouteIdent handler_route.go

// RouteIdent define type for route identifier.
type RouteIdent int

// Route identifiers.
const (
	RouteNone RouteIdent = iota
	RouteIncomplete
	RouteError
	RouteSuccess
	RouteToUniqueText
	RouteToUniqueJSON
)

var errEmptyComponentPart = errors.New("component part is empty")

func extractInt32R09(v string, index, bound int) (int32, int, error) {
	var result int32
	for idx := index; idx < bound; idx++ {
		ch := v[idx]
		if (ch >= '0') && (ch <= '9') {
			result = result*10 + int32(ch-'0')
		} else if idx == index {
			return result, idx, errEmptyComponentPart
		} else {
			return result, idx, nil
		}
	}
	return result, bound, nil
}

var errFragmentSmallerThanExpect = errors.New("remaining fragment smaller than expect")

func computeFragmentLiteralDigest(t string, digest uint64, index, bound, length int) (uint64, int, error) {
	b := index + length
	if b > bound {
		return digest, index, errFragmentSmallerThanExpect
	}
	for index < b {
		ch := t[index]
		index++
		digest = (digest << 8) | uint64(ch)
	}
	return digest, index, nil
}

func (h *sampleHandler) routeTermUT(w http.ResponseWriter, req *http.Request, index, bound int) (RouteIdent, error) {
	reqURI := req.RequestURI
	num, index, err := extractInt32R09(reqURI, index+5, bound)
	if nil != err {
		return RouteToUniqueText, err
	}
	h.uniqueText(w, req, index, bound, num)
	return RouteToUniqueText, nil
}

func (h *sampleHandler) routeTermUJ(w http.ResponseWriter, req *http.Request, index, bound int) (RouteIdent, error) {
	reqURI := req.RequestURI
	num, index, err := extractInt32R09(reqURI, index+5, bound)
	if nil != err {
		return RouteToUniqueJSON, err
	}
	h.uniqueJSON(w, req, index, bound, num)
	return RouteToUniqueJSON, nil
}

func (h *sampleHandler) routeViaU(w http.ResponseWriter, req *http.Request, index, bound int) (RouteIdent, error) {
	reqURI := req.RequestURI
	if index = index + 12; index >= bound {
		return RouteIncomplete, nil
	}
	switch reqURI[index] {
	case 'j':
		return h.routeTermUJ(w, req, index, bound)
	case 't':
		return h.routeTermUT(w, req, index, bound)
	}
	return RouteIncomplete, nil
}

func (h *sampleHandler) routeRequest(w http.ResponseWriter, req *http.Request) (RouteIdent, error) {
	reqURI := req.RequestURI
	index := 0
	bound := len(reqURI)
	for index < bound {
		if reqURI[index] == '/' {
			index++
			break
		}
		index++
	}
	if index >= bound {
		return RouteNone, nil
	}
	switch reqURI[index] {
	case 'u':
		return h.routeViaU(w, req, index, bound)
	}
	var digest uint64
	var err error
	if digest, index, err = computeFragmentLiteralDigest(reqURI, digest, index, bound, 5); nil != err {
		return RouteError, err
	} else if digest == 0x0000006465627567 {

	}

	return RouteNone, nil

}
