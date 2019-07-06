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

func extractStringR00(v string, index, bound int) (string, int, error) {
	var buf []byte
	for idx := index; idx < bound; idx++ {
		ch := v[idx]
		switch {
		case (ch >= 'a') && (ch <= 'z'):
			fallthrough
		case ch == '-':
			buf = append(buf, ch)
		default:
			return string(buf), idx, nil
		}
	}
	return string(buf), bound, nil
}

func extractStringBuiltInR01NoSlash(v string, offset, bound int) (string, int, error) {
	var buf []byte
	for idx := offset; idx < bound; idx++ {
		if ch := v[idx]; ch != '/' {
			buf = append(buf, ch)
			continue
		}
		return string(buf), idx, nil
	}
	return string(buf), bound, nil
}

var errFragmentSmallerThanExpect = errors.New("remaining path fragment smaller than expect")

func extractInt32BuiltInR01(v string, offset, bound int) (int32, int, error) {
	if bound <= offset {
		return 0, offset, errFragmentSmallerThanExpect
	}
	negative := false
	if ch := v[offset]; '-' == ch {
		negative = true
		offset++
	}
	var result int32
	for idx := offset; idx < bound; idx++ {
		ch := v[idx]
		digit := (ch & 0x0F)
		if ((ch & 0xF0) == 0x30) && ((1023 & (1 << digit)) != 0) {
			result = result*10 + int32(digit)
			continue
		}
		if negative {
			return -result, idx, nil
		}
		return result, idx, nil
	}
	if negative {
		return -result, bound, nil
	}
	return result, bound, nil
}

func extractUInt32BuiltInR01(v string, offset, bound int) (uint32, int, error) {
	if bound <= offset {
		return 0, offset, errFragmentSmallerThanExpect
	}
	var result uint32
	for idx := offset; idx < bound; idx++ {
		ch := v[idx]
		digit := (ch & 0x0F)
		if ((ch & 0xF0) == 0x30) && ((1023 & (1 << digit)) != 0) {
			result = result*10 + uint32(digit)
			continue
		}
		return result, idx, nil
	}
	return result, bound, nil
}

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

func computePrefixMatchingDigest32(path string, offset, bound, length int) (uint32, int, error) {
	b := offset + length
	if b > bound {
		return 0, offset, errFragmentSmallerThanExpect
	}
	var digest uint32
	for offset < b {
		ch := path[offset]
		offset++
		digest = (digest << 8) | uint32(ch)
	}
	return digest, offset, nil
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
	reqPath := req.URL.Path
	reqPathOffset := 0
	reqPathBound := len(reqPath)
	for reqPathOffset < reqPathBound {
		if reqPath[reqPathOffset] == '/' {
			reqPathOffset++
			break
		}
		reqPathOffset++
	}
	if reqPathOffset >= reqPathBound {
		return RouteNone, nil
	}
	switch reqPath[reqPathOffset] {
	case 'u':
		return h.routeViaU(w, req, reqPathOffset, reqPathBound)
	}
	var digest uint64
	var err error
	if digest, reqPathOffset, err = computeFragmentLiteralDigest(reqPath, digest, reqPathOffset, reqPathBound, 5); nil != err {
		return RouteError, err
	} else if digest == 0x0000006465627567 {

	}

	return RouteNone, nil

}
