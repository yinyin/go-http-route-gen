package httproutegen

import (
	"strconv"
	"strings"
)

func makeCodeTypeRouteIdent(routePrefix string) string {
	return "// " + (routePrefix + "RouteIdent") + " define type for route identifier.\n" +
		"type " + (routePrefix + "RouteIdent") + " int\n" +
		"\n"
}

func makeCodeConstRouteIdent(routePrefix string, targetHandlerRouteIdents []string) string {
	return "// Route identifiers.\n" +
		"const (\n" +
		"\t" + (routePrefix + "RouteNone " + routePrefix + "RouteIdent") + " = iota\n" +
		"\t" + (routePrefix + "Route") + "Incomplete\n" +
		"\t" + (routePrefix + "Route") + "Error\n" +
		"\t" + (routePrefix + "Route") + "Success\n" +
		(strings.Join(targetHandlerRouteIdents, "\n")) + "\n" +
		")\n" +
		"\n"
}

func makeCodeMethodRouteEnterance(routePrefix string, routingLogicCode string) string {
	return "func (h *localHandler) routeRequest(w http.ResponseWriter, req *http.Request) (" + (routePrefix + "RouteIdent") + ", error) {\n" +
		"\treqPath := req.URL.Path\n" +
		"\treqPathOffset := 0\n" +
		"\treqPathBound := len(reqPath)\n" +
		"\tfor reqPathOffset < reqPathBound {\n" +
		"\t\tif reqPath[reqPathOffset] == '/' {\n" +
		"\t\t\treqPathOffset++\n" +
		"\t\t\tbreak\n" +
		"\t\t}\n" +
		"\t\treqPathOffset++\n" +
		"\t}\n" +
		"\tif reqPathOffset >= reqPathBound {\n" +
		"\t\treturn " + (routePrefix + "RouteNone") + ", nil\n" +
		"\t}\n" +
		(routingLogicCode) + "\n" +
		"\treturn " + (routePrefix + "RouteNone") + ", nil\n" +
		"}\n" +
		"\n"
}

const codeFunctionComputePrefixMatching32 = "var errFragmentSmallerThanExpect = errors.New(\"remaining path fragment smaller than expect\")\n" +
	"\n" +
	"func computePrefixMatchingDigest32(path string, offset, bound, length int) (uint32, int, error) {\n" +
	"\tb := offset + length\n" +
	"\tif b > bound {\n" +
	"\t\treturn 0, offset, errFragmentSmallerThanExpect\n" +
	"\t}\n" +
	"\tvar digest uint32\n" +
	"\tfor offset < b {\n" +
	"\t\tch := path[offset]\n" +
	"\t\toffset++\n" +
	"\t\tdigest = (digest << 8) | uint32(ch)\n" +
	"\t}\n" +
	"\treturn digest, offset, nil\n" +
	"}\n" +
	"\n"

func makeCodeBlockPrefixMatching32Start(routePrefix string, baseOffset int, digestLength int) string {
	return "if digest32, reqPathOffset, err = computePrefixMatchingDigest32(reqPath, " + ("reqPathOffset" + codeTemplateGenIntPlus(baseOffset)) + ", reqPathBound, " + (strconv.FormatInt(int64(digestLength), 10)) + "); nil != err {\n" +
		"    return " + (routePrefix + "RouteError") + ", err\n" +
		"}\n" +
		"\n"
}

func makeCodeBlockPrefixMatching32Fork(routePrefix string, digestValue uint32, routingLogicCode string) string {
	return "else if digest32 == " + ("0x" + strconv.FormatInt(int64(digestValue), 16)) + " {\n" +
		(routingLogicCode) + "\n" +
		"}\n" +
		"\n"
}

func makeCodeBlockFuzzyMatchingU8Start(baseOffset int, fuzzyDepth int, fuzzyByteValue uint32, routingLogicCode string) string {
	return "if ch := reqPath[" + ("reqPathOffset" + codeTemplateGenIntPlus(baseOffset+fuzzyDepth)) + "]; ch == " + ("0x" + strconv.FormatInt(int64(fuzzyByteValue), 16)) + " {\n" +
		(routingLogicCode) + "\n" +
		"}\n" +
		"\n"
}

func makeCodeBlockFuzzyMatchingU16Start(baseOffset int, fuzzyDepth int, fuzzyByteValue uint32, routingLogicCode string) string {
	return "if ch := (uint16(reqPath[" + ("reqPathOffset" + codeTemplateGenIntPlus(baseOffset+fuzzyDepth-1)) + "]) << 8) | uint16(reqPath[" + ("reqPathOffset" + codeTemplateGenIntPlus(baseOffset+fuzzyDepth)) + "]); ch == " + ("0x" + strconv.FormatInt(int64(fuzzyByteValue), 16)) + " {\n" +
		(routingLogicCode) + "\n" +
		"}\n" +
		"\n"
}

func makeCodeBlockFuzzyMatchingU8U16Middle(fuzzyByteValue uint32, routingLogicCode string) string {
	return "else if ch == " + ("0x" + strconv.FormatInt(int64(fuzzyByteValue), 16)) + " {\n" +
		(routingLogicCode) + "\n" +
		"}\n" +
		"\n"
}
