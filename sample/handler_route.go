// Code generated by go-http-route-gen. DO NOT EDIT.

package main

import (
	"errors"
	"net/http"
)

// RouteIdent define type for route identifier.
type RouteIdent int

// Route identifiers.
const (
	RouteNone RouteIdent = iota
	RouteIncomplete
	RouteError
	RouteMissSampleAdmin
	RouteMissDebugSample
	RouteSuccess
	RouteToQueryProduct
	RouteToDownloadProduct
	RouteToListProducts
	RouteToShowProduct
	RouteToSampleData
	RouteToDebugText
	RouteToDebugJSON
	RouteToExactText
	RouteToDebugNumber
	RouteToUniqueText
	RouteToUniqueJSON
)

var errFragmentSmallerThanExpect = errors.New("remaining path fragment smaller than expect")

var filterMaskStringRxSeq000 = [...]uint32{0xfff01ff9, 0xfff03fff, 0x3fff, 0x0}

func extractStringRxSeq000(v string, offset, bound int) (string, int, error) {
	var result []byte
	for idx := offset; idx < bound; idx++ {
		ch := v[idx]
		moved := ch - 0x2d
		page := (moved >> 5) & 0x3
		nbit := moved & 0x1F
		if 0 != (filterMaskStringRxSeq000[page] & (1 << nbit)) {
			result = append(result, ch)
			continue
		}
		return string(result), idx, nil
	}
	return string(result), bound, nil
}

func extractInt64BuiltInR02(v string, offset, bound int) (int64, int, error) {
	if bound <= offset {
		return 0, offset, errFragmentSmallerThanExpect
	}
	var result int64
	for idx := offset; idx < bound; idx++ {
		ch := v[idx]
		digit := (ch & 0x0F)
		if ((ch & 0xF0) == 0x30) && ((1023 & (1 << digit)) != 0) {
			result = result*10 + int64(digit)
			continue
		}
		return result, idx, nil
	}
	return result, bound, nil
}

func extractInt32BuiltInR02(v string, offset, bound int) (int32, int, error) {
	if bound <= offset {
		return 0, offset, errFragmentSmallerThanExpect
	}
	var result int32
	for idx := offset; idx < bound; idx++ {
		ch := v[idx]
		digit := (ch & 0x0F)
		if ((ch & 0xF0) == 0x30) && ((1023 & (1 << digit)) != 0) {
			result = result*10 + int32(digit)
			continue
		}
		return result, idx, nil
	}
	return result, bound, nil
}

var filterMaskHexInt32BuiltInR03 = [...]uint16{0x7E, 0, 0x7E, 0x3FF}
var offsetValueHexInt32BuiltInR03 = [...]byte{9, 0, 9, 0}

func extractInt32BuiltInR03(v string, offset, bound int) (int32, int, error) {
	if bound <= offset {
		return 0, offset, errFragmentSmallerThanExpect
	}
	var result int32
	for idx := offset; idx < bound; idx++ {
		ch := v[idx]
		digit := (ch & 0x0F)
		page := ((ch >> 4) & 0x3)
		if (filterMaskHexInt32BuiltInR03[page] & (1 << digit)) != 0 {
			result = result<<4 | int32(digit+offsetValueHexInt32BuiltInR03[page])
			continue
		}
		return result, idx, nil
	}
	return result, bound, nil
}

func extractUInt32BuiltInR03(v string, offset, bound int) (uint32, int, error) {
	if bound <= offset {
		return 0, offset, errFragmentSmallerThanExpect
	}
	var result uint32
	for idx := offset; idx < bound; idx++ {
		ch := v[idx]
		digit := (ch & 0x0F)
		page := ((ch >> 4) & 0x3)
		if (filterMaskHexInt32BuiltInR03[page] & (1 << digit)) != 0 {
			result = result<<4 | uint32(digit+offsetValueHexInt32BuiltInR03[page])
			continue
		}
		return result, idx, nil
	}
	return result, bound, nil
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
	var err error
	var digest32 uint32
	if digest32, reqPathOffset, err = computePrefixMatchingDigest32(reqPath, reqPathOffset, reqPathBound, 4); nil != err {
		return RouteError, err
	} else if digest32 == 0x73616d70 {
		if digest32, reqPathOffset, err = computePrefixMatchingDigest32(reqPath, reqPathOffset, reqPathBound, 4); nil != err {
			return RouteError, err
		} else if digest32 == 0x6c652d61 {
			if reqPathOffset = reqPathOffset + 3; reqPathOffset >= reqPathBound {
				return RouteIncomplete, nil
			}
			if ch := reqPath[reqPathOffset]; ch == 0x71 {
				var productName string
				if productName, reqPathOffset, err = extractStringRxSeq000(reqPath, reqPathOffset+6, reqPathBound); nil != err {
					return RouteError, err
				}
				switch req.Method {
				case http.MethodGet:
					fallthrough
				case http.MethodPost:
					h.queryProduct(w, req, reqPathOffset, productName)
					return RouteToQueryProduct, nil
				}
				http.Error(w, "not allow", http.StatusMethodNotAllowed)
				return RouteError, nil
			} else if ch == 0x64 {
				var sessionId int64
				if sessionId, reqPathOffset, err = extractInt64BuiltInR02(reqPath, reqPathOffset+9, reqPathBound); nil != err {
					return RouteError, err
				}
				var targetId int64
				if targetId, reqPathOffset, err = extractInt64BuiltInR02(reqPath, reqPathOffset+1, reqPathBound); nil != err {
					return RouteError, err
				}
				switch req.Method {
				case http.MethodGet:
					h.downloadProduct(w, req, reqPathOffset, sessionId, targetId)
					return RouteToDownloadProduct, nil
				}
				http.Error(w, "not allow", http.StatusMethodNotAllowed)
				return RouteError, nil
			} else if ch == 0x6e {
				if reqPathOffset = reqPathOffset + 13; reqPathOffset >= reqPathBound {
					return RouteMissSampleAdmin, nil
				}
				if ch := reqPath[reqPathOffset]; ch == 0x73 {
					switch req.Method {
					case http.MethodGet:
						h.listProducts(w, req, reqPathOffset+1)
						return RouteToListProducts, nil
					}
					http.Error(w, "not allow", http.StatusMethodNotAllowed)
					return RouteError, nil
				} else if ch == 0x2f {
					var productId int64
					if productId, reqPathOffset, err = extractInt64BuiltInR02(reqPath, reqPathOffset+1, reqPathBound); nil != err {
						return RouteMissSampleAdmin, err
					}
					switch req.Method {
					case http.MethodGet:
						h.showProduct(w, req, reqPathOffset, productId)
						return RouteToShowProduct, nil
					}
					http.Error(w, "not allow", http.StatusMethodNotAllowed)
					return RouteError, nil
				}
				return RouteMissSampleAdmin, nil
			}
		} else if digest32 == 0x6c652d64 {
			if digest32, reqPathOffset, err = computePrefixMatchingDigest32(reqPath, reqPathOffset, reqPathBound, 3); nil != err {
				return RouteError, err
			} else if digest32 == 0x617461 {
				switch req.Method {
				case http.MethodGet:
					h.sampleData(w, req, reqPathOffset)
					return RouteToSampleData, nil
				}
				http.Error(w, "not allow", http.StatusMethodNotAllowed)
				return RouteError, nil
			} else if digest32 == 0x656275 {
				if reqPathOffset = reqPathOffset + 2; reqPathOffset >= reqPathBound {
					return RouteIncomplete, nil
				}
				if ch := reqPath[reqPathOffset]; ch == 0x74 {
					switch req.Method {
					case http.MethodGet:
						h.debugText(w, req, reqPathOffset+4)
						return RouteToDebugText, nil
					}
					http.Error(w, "not allow", http.StatusMethodNotAllowed)
					return RouteError, nil
				} else if ch == 0x6a {
					switch req.Method {
					case http.MethodGet:
						h.debugJSON(w, req, reqPathOffset+4)
						return RouteToDebugJSON, nil
					}
					http.Error(w, "not allow", http.StatusMethodNotAllowed)
					return RouteError, nil
				}
			}
		} else if digest32 == 0x6c652d65 {
			if digest32, reqPathOffset, err = computePrefixMatchingDigest32(reqPath, reqPathOffset, reqPathBound, 4); nil != err {
				return RouteError, err
			} else if digest32 == 0x78616374 {
				if digest32, reqPathOffset, err = computePrefixMatchingDigest32(reqPath, reqPathOffset, reqPathBound, 4); nil != err {
					return RouteError, err
				} else if digest32 == 0x2f746578 {
					if digest32, reqPathOffset, err = computePrefixMatchingDigest32(reqPath, reqPathOffset, reqPathBound, 1); nil != err {
						return RouteError, err
					} else if digest32 == 0x74 {
						switch req.Method {
						case http.MethodGet:
							h.exactText(w, req, reqPathOffset)
							return RouteToExactText, nil
						}
						http.Error(w, "not allow", http.StatusMethodNotAllowed)
						return RouteError, nil
					}
				}
			}
		}
	} else if digest32 == 0x64656275 {
		if digest32, reqPathOffset, err = computePrefixMatchingDigest32(reqPath, reqPathOffset, reqPathBound, 1); nil != err {
			return RouteMissDebugSample, err
		} else if digest32 == 0x67 {
			var num int32
			if num, reqPathOffset, err = extractInt32BuiltInR02(reqPath, reqPathOffset+14, reqPathBound); nil != err {
				return RouteMissDebugSample, err
			}
			var hex1 int32
			if hex1, reqPathOffset, err = extractInt32BuiltInR03(reqPath, reqPathOffset+2, reqPathBound); nil != err {
				return RouteMissDebugSample, err
			}
			var hex2 uint32
			if hex2, reqPathOffset, err = extractUInt32BuiltInR03(reqPath, reqPathOffset+1, reqPathBound); nil != err {
				return RouteMissDebugSample, err
			}
			switch req.Method {
			case http.MethodGet:
				h.debugNumber(w, req, reqPathOffset, num, hex1, hex2)
				return RouteToDebugNumber, nil
			}
			http.Error(w, "not allow", http.StatusMethodNotAllowed)
			return RouteError, nil
		}
		return RouteMissDebugSample, nil
	} else if digest32 == 0x756e6971 {
		if reqPathOffset = reqPathOffset + 8; reqPathOffset >= reqPathBound {
			return RouteIncomplete, nil
		}
		if ch := reqPath[reqPathOffset]; ch == 0x74 {
			var num int32
			if num, reqPathOffset, err = extractInt32BuiltInR02(reqPath, reqPathOffset+5, reqPathBound); nil != err {
				return RouteError, err
			}
			switch req.Method {
			case http.MethodGet:
				h.uniqueText(w, req, reqPathOffset, num)
				return RouteToUniqueText, nil
			}
			http.Error(w, "not allow", http.StatusMethodNotAllowed)
			return RouteError, nil
		} else if ch == 0x6a {
			var num int32
			if num, reqPathOffset, err = extractInt32BuiltInR02(reqPath, reqPathOffset+5, reqPathBound); nil != err {
				return RouteError, err
			}
			switch req.Method {
			case http.MethodGet:
				h.uniqueJSON(w, req, reqPathOffset, num)
				return RouteToUniqueJSON, nil
			}
			http.Error(w, "not allow", http.StatusMethodNotAllowed)
			return RouteError, nil
		}
	}
	return RouteNone, nil
}
