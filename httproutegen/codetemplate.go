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

func makeCodeMethodRouteEnterance(routePrefix string, receiverName string, handlerTypeName string, routeMethodName string, routingLogicCode string) string {
	return "func (" + (receiverName) + " *" + (handlerTypeName) + ") " + (routeMethodName) + "(w http.ResponseWriter, req *http.Request) (" + (routePrefix + "RouteIdent") + ", error) {\n" +
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

const codeErrFragmentSmallerThanExpect = "var errFragmentSmallerThanExpect = errors.New(\"remaining path fragment smaller than expect\")\n" +
	"\n"

const codeFunctionComputePrefixMatching32 = "func computePrefixMatchingDigest32(path string, offset, bound, length int) (uint32, int, error) {\n" +
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

func makeCodeBlockGetParameter(routePrefix string, paramName string, paramType string, extractFuncName string, baseOffset int, routingLogicCode string) string {
	return "var " + (paramName) + " " + (paramType) + "\n" +
		"if " + (paramName) + ", reqPathOffset, err = " + (extractFuncName) + "(reqPath, " + ("reqPathOffset" + codeTemplateGenIntPlus(baseOffset)) + ", reqPathBound); nil != err {\n" +
		"    return " + (routePrefix + "RouteError") + ", err\n" +
		"}\n" +
		(routingLogicCode) + "\n" +
		"\n"
}

const codeMethodExtractStringBuiltInR01NoSlash = "func extractStringBuiltInR01NoSlash(v string, offset, bound int) (string, int, error) {\n" +
	"\tvar buf []byte\n" +
	"\tfor idx := offset; idx < bound; idx++ {\n" +
	"\t\tif ch := v[idx]; ch != '/' {\n" +
	"\t\t\tbuf = append(buf, ch)\n" +
	"\t\t\tcontinue\n" +
	"\t\t}\n" +
	"\t\treturn string(buf), idx, nil\n" +
	"\t}\n" +
	"\treturn string(buf), bound, nil\n" +
	"}\n" +
	"\n"

func makeCodeMethodExtractIntBuiltInR01(typeBit string) string {
	return "func extractInt" + (typeBit) + "BuiltInR01(v string, offset, bound int) (int" + (typeBit) + ", int, error) {\n" +
		"\tif bound <= offset {\n" +
		"\t\treturn 0, offset, errFragmentSmallerThanExpect\n" +
		"\t}\n" +
		"\tnegative := false\n" +
		"\tif ch := v[offset]; '-' == ch {\n" +
		"\t\tnegative = true\n" +
		"\t\toffset++\n" +
		"\t}\n" +
		"\tvar result int" + (typeBit) + "\n" +
		"\tfor idx := offset; idx < bound; idx++ {\n" +
		"\t\tch := v[idx]\n" +
		"\t\tdigit := (ch & 0x0F)\n" +
		"\t\tif ((ch & 0xF0) == 0x30) && ((1023 & (1 << digit)) != 0) {\n" +
		"\t\t\tresult = result*10 + int" + (typeBit) + "(digit)\n" +
		"\t\t\tcontinue\n" +
		"\t\t}\n" +
		"\t\tif negative {\n" +
		"\t\t\treturn -result, idx, nil\n" +
		"\t\t}\n" +
		"\t\treturn result, idx, nil\n" +
		"\t}\n" +
		"\tif negative {\n" +
		"\t\treturn -result, bound, nil\n" +
		"\t}\n" +
		"\treturn result, bound, nil\n" +
		"}\n" +
		"\n"
}

func makeCodeMethodExtractUIntBuiltInR02(typeTitle string, typeName string) string {
	return "func extract" + (typeTitle) + "BuiltInR02(v string, offset, bound int) (" + (typeName) + ", int, error) {\n" +
		"\tif bound <= offset {\n" +
		"\t\treturn 0, offset, errFragmentSmallerThanExpect\n" +
		"\t}\n" +
		"\tvar result " + (typeName) + "\n" +
		"\tfor idx := offset; idx < bound; idx++ {\n" +
		"\t\tch := v[idx]\n" +
		"\t\tdigit := (ch & 0x0F)\n" +
		"\t\tif ((ch & 0xF0) == 0x30) && ((1023 & (1 << digit)) != 0) {\n" +
		"\t\t\tresult = result*10 + " + (typeName) + "(digit)\n" +
		"\t\t\tcontinue\n" +
		"\t\t}\n" +
		"\t\treturn result, idx, nil\n" +
		"\t}\n" +
		"\treturn result, bound, nil\n" +
		"}\n" +
		"\n"
}
