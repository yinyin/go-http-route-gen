# Heading Code

* `keep-empty-line`

```go
package httproutegen

import (
	"strconv"
	"strings"
)

```

# Route Ident

* `builder`: `makeCodeTypeRouteIdent`, `routePrefix string`
* `preserve-new-line`
* `replace`:
  - ``` (RouteIdent) ```
  - `$1`
  - ``` routePrefix + "RouteIdent" ```

```go
// RouteIdent define type for route identifier.
type RouteIdent int
```

# Default Route Idents

* `builder`: `makeCodeConstRouteIdent`, `routePrefix string`, `coveredAreaRouteIdents []string`, `targetHandlerRouteIdents []string`
* `preserve-new-line`
* `replace`:
  - ``` (\s*RouteMissingCoveredArea) ```
  - `$1`
  - ``` strings.Join(coveredAreaRouteIdents, "\n") ```
* `replace`:
  - ``` (\s*RouteToTargetHandler) ```
  - `$1`
  - ``` strings.Join(targetHandlerRouteIdents, "\n") ```
* `replace`:
  - ``` (RouteNone RouteIdent) ```
  - `$1`
  - ``` routePrefix + "RouteNone " + routePrefix + "RouteIdent" ```
* `replace`:
  - ``` (Route)[A-Z] ```
  - `$1`
  - ``` routePrefix + "Route" ```

```go
// Route identifiers.
const (
	RouteNone RouteIdent = iota
	RouteIncomplete
    RouteError
    RouteMissingCoveredArea
	RouteSuccess
	RouteToTargetHandler
)
```

# Route Method

* `builder`: `makeCodeMethodRouteEnterance`, `routePrefix string`, `receiverName string`, `handlerTypeName string`, `routeMethodName string`, `routingLogicCode string`
* `preserve-new-line`
* `replace`:
  - ``` \((h) \*(localHandler)\) (routeRequest)\( ```
  - `$1`
  - ``` receiverName ```
  - `$2`
  - ``` handlerTypeName ```
  - `$3`
  - ``` routeMethodName ```
* `replace`:
  - ``` (RouteIdent) ```
  - `$1`
  - ``` routePrefix + "RouteIdent" ```
* `replace`:
  - ``` (RouteNone) ```
  - `$1`
  - ``` routePrefix + "RouteNone" ```
* `replace`:
  - ``` (\s*InvokeRoutingLogic\(\)) ```
  - `$1`
  - ``` routingLogicCode ```

```go
func (h *localHandler) routeRequest(w http.ResponseWriter, req *http.Request) (RouteIdent, error) {
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
	InvokeRoutingLogic()
	return RouteNone, nil
}
```

# Error (errFragmentSmallerThanExpect)

* `const`: `codeErrFragmentSmallerThanExpect`
* `preserve-new-line`

```go
var errFragmentSmallerThanExpect = errors.New("remaining path fragment smaller than expect")
```

# Compute Prefix Matching Digest Value (UINT-32)

* `const`: `codeFunctionComputePrefixMatching32`
* `preserve-new-line`

```go
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
```

# Code of Prefix Matching Logic (Start)

* `builder`: `makeCodeBlockPrefixMatching32Start`, `routePrefix string`, `routeMissingIdent string`, `baseOffset int`, `digestLength int`
* `preserve-new-line`
* `replace`:
  - ``` reqPath, (reqPathOffset), reqPathBound ```
  - `$1`
  - ``` "reqPathOffset" + codeTemplateGenIntPlus(baseOffset) ```
* `replace`:
  - ``` (RouteError) ```
  - `$1`
  - ``` pickNonEmptyIdent(routeMissingIdent, routePrefix + "RouteError") ```
* `replace`:
  - ``` (DigestLen) ```
  - `$1`
  - ``` strconv.FormatInt(int64(digestLength), 10) ```

```go
if digest32, reqPathOffset, err = computePrefixMatchingDigest32(reqPath, reqPathOffset, reqPathBound, DigestLen); nil != err {
	return RouteError, err
}
```

# Code of Prefix Matching Logic (Fork)

* `builder`: `makeCodeBlockPrefixMatching32Fork`, `routePrefix string`, `digestValue uint32`, `routingLogicCode string`
* `preserve-new-line`
* `replace`:
  - ``` (DigestValue) ```
  - `$1`
  - ``` "0x" + strconv.FormatInt(int64(digestValue), 16) ```
* `replace`:
  - ``` (\s*InvokeRoutingLogic\(\)) ```
  - `$1`
  - ``` routingLogicCode ```

```go
else if digest32 == DigestValue {
	InvokeRoutingLogic()
}
```

# Code of Fuzzy Matching Logic (Boundary Check, Non-zero)

* `builder`: `makeCodeBlockFuzzyMatchingBoundCheckNonZero`, `routePrefix string`, `routeMissingIdent string`, `baseOffset int`, `fuzzyDepth int`
* `preserve-new-line`
* `replace`:
  - ``` reqPathOffset = (reqPathOffset \+ 3) ```
  - `$1`
  - ``` "reqPathOffset" + codeTemplateGenIntPlus(baseOffset+fuzzyDepth) ```
* `replace`:
  - ``` (RouteIncomplete) ```
  - `$1`
  - ``` pickNonEmptyIdent(routeMissingIdent, routePrefix + "RouteIncomplete") ```

```go
if reqPathOffset = reqPathOffset + 3; reqPathOffset >= reqPathBound {
	return RouteIncomplete, nil
}
```

# Code of Fuzzy Matching Logic (Boundary Check, Zero)

* `builder`: `makeCodeBlockFuzzyMatchingBoundCheckZero`, `routePrefix string`, `routeMissingIdent string`
* `preserve-new-line`
* `replace`:
  - ``` (RouteIncomplete) ```
  - `$1`
  - ``` pickNonEmptyIdent(routeMissingIdent, routePrefix + "RouteIncomplete") ```

```go
if reqPathOffset >= reqPathBound {
	return RouteIncomplete, nil
}
```

# Code of Fuzzy Matching Logic (U8, Start)

* `builder`: `makeCodeBlockFuzzyMatchingU8Start`, `fuzzyByteValue uint32`, `routingLogicCode string`
* `preserve-new-line`
* `replace`:
  - ``` == (FuzzyByte) ```
  - `$1`
  - ``` "0x" + strconv.FormatInt(int64(fuzzyByteValue), 16) ```
* `replace`:
  - ``` (\s*InvokeRoutingLogic\(\)) ```
  - `$1`
  - ``` routingLogicCode ```

```go
if ch := reqPath[reqPathOffset]; ch == FuzzyByte {
	InvokeRoutingLogic()
}
```

# Code of Fuzzy Matching Logic (U16, Start)

* `builder`: `makeCodeBlockFuzzyMatchingU16Start`, `fuzzyByteValue uint32`, `routingLogicCode string`
* `preserve-new-line`
* `replace`:
  - ``` == (FuzzyByte) ```
  - `$1`
  - ``` "0x" + strconv.FormatInt(int64(fuzzyByteValue), 16) ```
* `replace`:
  - ``` (\s*InvokeRoutingLogic\(\)) ```
  - `$1`
  - ``` routingLogicCode ```

```go
if ch := (uint16(reqPath[reqPathOffset-1]) << 8) | uint16(reqPath[reqPathOffset]); ch == FuzzyByte {
	InvokeRoutingLogic()
}
```

# Code of Fuzzy Matching Logic (U8, U16, Middle)

* `builder`: `makeCodeBlockFuzzyMatchingU8U16Middle`, `fuzzyByteValue uint32`, `routingLogicCode string`
* `preserve-new-line`
* `replace`:
  - ``` == (FuzzyByte) ```
  - `$1`
  - ``` "0x" + strconv.FormatInt(int64(fuzzyByteValue), 16) ```
* `replace`:
  - ``` (\s*InvokeRoutingLogic\(\)) ```
  - `$1`
  - ``` routingLogicCode ```

```go
else if ch == FuzzyByte {
	InvokeRoutingLogic()
}
```

# Get Parameter

* `builder`: `makeCodeBlockGetParameter`, `routePrefix string`, `routeMissingIdent string`, `paramName string`, `paramType string`, `extractFuncName string`, `baseOffset int`, `routingLogicCode string`
* `preserve-new-line`
* `replace`:
  - ``` var (paramName) (string) ```
  - `$1`
  - ``` paramName ```
  - `$2`
  - ``` paramType ```
* `replace`:
  - ``` if (paramName), reqPathOffset, err = (extractParameterFunction)\(reqPath, (reqPathOffset), reqPathBound ```
  - `$1`
  - ``` paramName ```
  - `$2`
  - ``` extractFuncName ```
  - `$3`
  - ``` "reqPathOffset" + codeTemplateGenIntPlus(baseOffset) ```
* `replace`:
  - ``` (RouteError) ```
  - `$1`
  - ``` pickNonEmptyIdent(routeMissingIdent, routePrefix + "RouteError") ```
* `replace`:
  - ``` (\s*InvokeRoutingLogic\(\)) ```
  - `$1`
  - ``` routingLogicCode ```

```go
var paramName string
if paramName, reqPathOffset, err = extractParameterFunction(reqPath, reqPathOffset, reqPathBound); nil != err {
	return RouteError, err
}
InvokeRoutingLogic()
```

# Invoke without Match Method

* `builder`: `makeCodeBlockNoMatchMethodForInvoke`, `routePrefix string`
* `preserve-new-line`
* `replace`:
  - ``` (RouteError) ```
  - `$1`
  - ``` (routePrefix + "RouteError") ```

```go
http.Error(w, "not allow", http.StatusMethodNotAllowed)
return RouteError, nil
```


# Extract Function (^\ => string, no-converter)

* `const`: `codeMethodExtractStringBuiltInR01NoSlash`
* `preserve-new-line`

```go
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
```

# Extract Function (0-9\- => signed int32/64, no-converter)

* `builder`: `makeCodeMethodExtractIntBuiltInR01`, `typeBit string`
* `preserve-new-line`
* `replace`:
  - ``` extractInt(32)BuiltInR01\(v string, offset, bound int\) \(int(32), int, error\) ```
  - `$1`
  - ``` typeBit ```
  - `$2`
  - ``` typeBit ```
* `replace`:
  - ``` var result int(32) ```
  - `$1`
  - ``` typeBit ```
* `replace`:
  - ``` int(32)\(digit\) ```
  - `$1`
  - ``` typeBit ```

```go
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
```

# Extract Function (0-9 => unsigned int32/64, no-converter)

* `builder`: `makeCodeMethodExtractUIntBuiltInR02`, `typeTitle string`, `typeName string`
* `preserve-new-line`
* `replace`:
  - ``` extract(UInt32)BuiltInR02\(v string, offset, bound int\) \((uint32), int, error\) ```
  - `$1`
  - ``` typeTitle ```
  - `$2`
  - ``` typeName ```
* `replace`:
  - ``` var result (uint32) ```
  - `$1`
  - ``` typeName ```
* `replace`:
  - ``` (uint32)\(digit\) ```
  - `$1`
  - ``` typeName ```

```go
func extractUInt32BuiltInR02(v string, offset, bound int) (uint32, int, error) {
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
```

# Extract Function Support Constants (0-9A-Fa-f => signed/unsigned int32/64, no-converter)

* `const`: `codeSupportConstantsExtractHexIntBuiltInR03`
* `preserve-new-line`

```go
var filterMaskHexInt32BuiltInR03 = [...]uint16{0x7E, 0, 0x7E, 0x3FF}
var offsetValueHexInt32BuiltInR03 = [...]byte{9, 0, 9, 0}
```

# Extract Function (0-9A-Fa-f => signed/unsigned int32/64, no-converter)

* `builder`: `makeCodeMethodExtractHexIntBuiltInR03`, `typeTitle string`, `typeName string`
* `preserve-new-line`
* `replace`:
  - ``` extract(Int32)BuiltInR03\(v string, offset, bound int\) \((int32), int, error\) ```
  - `$1`
  - ``` typeTitle ```
  - `$2`
  - ``` typeName ```
* `replace`:
  - ``` var result (int32) ```
  - `$1`
  - ``` typeName ```
* `replace`:
  - ``` (int32)\(digit ```
  - `$1`
  - ``` typeName ```

```go
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
```

# Extract Function (bit-map => []byte/string, no-converter)

* `builder`: `makeCodeMethodExtractByteSliceStringBitMasked`, `typeTitle string`, `typeName string`, `typeCasting string`, `rangeBase byte`, `bitmaskIdent string`, `bitmaskSlice []uint32`
* `preserve-new-line`
* `replace`:
  - ``` filterMask(String)Rx(00000000) = \[\.\.\.\]uint32{0x(0), 0x(1), 0x(2), 0x(3)} ```
  - `$1`
  - ``` typeTitle ```
  - `$2`
  - ``` bitmaskIdent ```
  - `$3`
  - ``` strconv.FormatInt(int64(bitmaskSlice[0]), 16) ```
  - `$4`
  - ``` strconv.FormatInt(int64(bitmaskSlice[1]), 16) ```
  - `$5`
  - ``` strconv.FormatInt(int64(bitmaskSlice[2]), 16) ```
  - `$6`
  - ``` strconv.FormatInt(int64(bitmaskSlice[3]), 16) ```
* `replace`:
  - ``` extract(String)Rx(00000000)\(v string, offset, bound int\) \((string), int, error\) ```
  - `$1`
  - ``` typeTitle ```
  - `$2`
  - ``` bitmaskIdent ```
  - `$3`
  - ``` typeName ```
* `replace`:
  - ``` moved := ch - (generalBase) ```
  - `$1`
  - ``` "0x" + strconv.FormatInt(int64(rangeBase), 16) ```
* `replace`:
  - ``` filterMask(String)Rx(00000000)\[page\] ```
  - `$1`
  - ``` typeTitle ```
  - `$2`
  - ``` bitmaskIdent ```
* `replace`:
  - ``` return (string)\(result\), ```
  - `$1`
  - ``` typeCasting ```

```go
var filterMaskStringRx00000000 = [...]uint32{0x0, 0x1, 0x2, 0x3}

func extractStringRx00000000(v string, offset, bound int) (string, int, error) {
	var result []byte
	for idx := offset; idx < bound; idx++ {
		ch := v[idx]
		moved := ch - generalBase
		page := (moved >> 5) & 0x3
		nbit := moved & 0x1F
		if 0 != (filterMaskStringRx00000000[page] & (1 << nbit)) {
			result = append(result, ch)
			continue
		}
		return string(result), idx, nil
	}
	return string(result), bound, nil
}
```
