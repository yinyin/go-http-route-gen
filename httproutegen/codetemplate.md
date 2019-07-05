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

* `builder`: `makeCodeConstRouteIdent`, `routePrefix string`, `targetHandlerRouteIdents []string`
* `preserve-new-line`
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
	RouteSuccess
	RouteToTargetHandler
)
```

# Route Method

* `builder`: `makeCodeMethodRouteEnterance`, `routePrefix string`, `routingLogicCode string`
* `preserve-new-line`
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
    InvokeRoutingLogic()
	return RouteNone, nil
}
```

# Compute Prefix Matching Digest Value (UINT-32)

* `const`: `codeFunctionComputePrefixMatching32`
* `preserve-new-line`

```go
var errFragmentSmallerThanExpect = errors.New("remaining path fragment smaller than expect")

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

* `builder`: `makeCodeBlockPrefixMatching32Start`, `routePrefix string`, `baseOffset int`, `digestLength int`
* `preserve-new-line`
* `replace`:
  - ``` reqPath, (reqPathOffset), reqPathBound ```
  - `$1`
  - ``` "reqPathOffset" + codeTemplateGenIntPlus(baseOffset) ```
* `replace`:
  - ``` (RouteError) ```
  - `$1`
  - ``` routePrefix + "RouteError" ```
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

# Code of Fuzzy Matching Logic (U8, Start)

* `builder`: `makeCodeBlockFuzzyMatchingU8Start`, `baseOffset int`, `fuzzyDepth int`, `fuzzyByteValue uint32`, `routingLogicCode string`
* `preserve-new-line`
* `replace`:
  - ``` \[(reqPathOffset)\] ```
  - `$1`
  - ``` "reqPathOffset" + codeTemplateGenIntPlus(baseOffset+fuzzyDepth) ```
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

* `builder`: `makeCodeBlockFuzzyMatchingU16Start`, `baseOffset int`, `fuzzyDepth int`, `fuzzyByteValue uint32`, `routingLogicCode string`
* `preserve-new-line`
* `replace`:
  - ``` \[(reqPathOffset0)\] ```
  - `$1`
  - ``` "reqPathOffset" + codeTemplateGenIntPlus(baseOffset+fuzzyDepth-1) ```
* `replace`:
  - ``` \[(reqPathOffset1)\] ```
  - `$1`
  - ``` "reqPathOffset" + codeTemplateGenIntPlus(baseOffset+fuzzyDepth) ```
* `replace`:
  - ``` == (FuzzyByte) ```
  - `$1`
  - ``` "0x" + strconv.FormatInt(int64(fuzzyByteValue), 16) ```
* `replace`:
  - ``` (\s*InvokeRoutingLogic\(\)) ```
  - `$1`
  - ``` routingLogicCode ```

```go
if ch := (uint16(reqPath[reqPathOffset0]) << 8) | uint16(reqPath[reqPathOffset1]); ch == FuzzyByte {
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
