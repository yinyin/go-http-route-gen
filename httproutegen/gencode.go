package httproutegen

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

func codeTemplateGenIntPlus(v int) string {
	if v == 0 {
		return ""
	}
	return fmt.Sprintf("+ %d", v)
}

func cleanupCodeBlock(codeText string, indent bool) string {
	codeLines := strings.Split(codeText, "\n")
	var c []string
	for _, l := range codeLines {
		l = strings.TrimRightFunc(l, unicode.IsSpace)
		if l == "" {
			continue
		}
		if indent {
			l = "\t" + l
		}
		c = append(c, l)
	}
	return strings.Join(c, "\n")
}

// CodeGenerateInstance keep variables for generating code.
type CodeGenerateInstance struct {
	fp             *os.File
	rootFanoutFork *FanoutFork

	PackageName     string
	ReceiverName    string
	HandlerTypeName string
	NamePrefix      string

	ImportModules []string
	HandlerNames  []string

	UsePrefixMatching bool
}

// OpenCodeGenerateInstance create an instance of code generator
func OpenCodeGenerateInstance(codeFilePath string, rootFanoutFork *FanoutFork) (inst *CodeGenerateInstance, err error) {
	fp, err := os.Create(codeFilePath)
	if nil != err {
		return
	}
	inst = &CodeGenerateInstance{
		fp:             fp,
		rootFanoutFork: rootFanoutFork,
	}
	inst.hasPrefixMatching(rootFanoutFork)
	inst.collectHandlerNames(rootFanoutFork)
	inst.addImportModule("net/http", false)
	return
}

// Close release allocated resources.
func (inst *CodeGenerateInstance) Close() (err error) {
	fp := inst.fp
	inst.fp = nil
	return fp.Close()
}

// addImportModule must invoke before `Generate()` code.
func (inst *CodeGenerateInstance) addImportModule(moduleName string, escaped bool) {
	if !escaped {
		moduleName = strconv.Quote(moduleName)
	}
	for _, modName := range inst.ImportModules {
		if modName == moduleName {
			return
		}
	}
	inst.ImportModules = append(inst.ImportModules, moduleName)
}

// addHandlerName must invoke before `Generate()` code.
func (inst *CodeGenerateInstance) addHandlerName(handlerName string) {
	for _, hndName := range inst.HandlerNames {
		if hndName == handlerName {
			return
		}
	}
	inst.HandlerNames = append(inst.HandlerNames, handlerName)
}

func (inst *CodeGenerateInstance) validateConfiguration() (err error) {
	if "" == inst.PackageName {
		return errors.New("package name is required")
	}
	return nil
}

func (inst *CodeGenerateInstance) hasPrefixMatching(fanoutFork *FanoutFork) {
	if fanoutFork.LogicType == LogicTypePrefixMatching {
		inst.UsePrefixMatching = true
		inst.addImportModule("errors", false)
		return
	}
	for _, childFork := range fanoutFork.ChildForks {
		inst.hasPrefixMatching(childFork)
	}
}

func (inst *CodeGenerateInstance) collectHandlerNames(fanoutFork *FanoutFork) {
	if fanoutFork.LogicType == LogicTypeInvokeHandler {
		inst.addHandlerName(fanoutFork.InvokeHandlerFanout.Route.HandlerName)
		return
	}
	for _, childFork := range fanoutFork.ChildForks {
		inst.collectHandlerNames(childFork)
	}
}

func (inst *CodeGenerateInstance) makeRouteIdentName(handlerName string) string {
	hnd := []rune(handlerName)
	hnd[0] = unicode.ToTitle(hnd[0])
	return inst.NamePrefix + "RouteTo" + string(hnd)
}

func (inst *CodeGenerateInstance) generateRouteIdentDefinitionListCode() string {
	var routeIdentNames []string
	for _, handlerName := range inst.HandlerNames {
		routeIdentNames = append(routeIdentNames, inst.makeRouteIdentName(handlerName))
	}
	return makeCodeConstRouteIdent(inst.NamePrefix, routeIdentNames)
}

func (inst *CodeGenerateInstance) generateSubForkFanoutCode(fanoutFork *FanoutFork, terminateSerials []int32) (result string) {
	subForks := fanoutFork.FindChildForkViaTerminateSerials(terminateSerials)
	for _, subFork := range subForks {
		if len(subForks) > 1 {
			result += "// WARN: multiple sub-forks.\n"
		}
		result += cleanupCodeBlock(inst.generateFanoutCode(subFork), true)
	}
	if "" == result {
		result = fmt.Sprintf("// WARN: empty sub-fork routing code: %v.\n", terminateSerials)
	}
	return
}

func (inst *CodeGenerateInstance) generatePrefixMatching(fanoutFork *FanoutFork) (result string) {
	result = makeCodeBlockPrefixMatching32Start(inst.NamePrefix, fanoutFork.BaseOffset, fanoutFork.PrefixLiteralDigests.Depth)
	for _, digestSet := range fanoutFork.PrefixLiteralDigests.Digests {
		subRoutingCode := inst.generateSubForkFanoutCode(fanoutFork, digestSet.TerminateSerials)
		codeText := makeCodeBlockPrefixMatching32Fork(inst.NamePrefix, digestSet.Value, subRoutingCode)
		result += codeText
	}
	return
}

func (inst *CodeGenerateInstance) generateFuzzyMatchingU8(fanoutFork *FanoutFork) (result string) {
	for idx, trackSet := range fanoutFork.FuzzyTracker.BestU8 {
		subRoutingCode := inst.generateSubForkFanoutCode(fanoutFork, trackSet.TerminateSerials)
		if idx == 0 {
			result += makeCodeBlockFuzzyMatchingU8Start(fanoutFork.BaseOffset, fanoutFork.FuzzyTracker.BestU8Depth, trackSet.Value, subRoutingCode)
		} else {
			result += makeCodeBlockFuzzyMatchingU8U16Middle(trackSet.Value, subRoutingCode)
		}
	}
	return
}

func (inst *CodeGenerateInstance) generateFuzzyMatchingU16(fanoutFork *FanoutFork) (result string) {
	for idx, trackSet := range fanoutFork.FuzzyTracker.BestU16 {
		subRoutingCode := inst.generateSubForkFanoutCode(fanoutFork, trackSet.TerminateSerials)
		if idx == 0 {
			result += makeCodeBlockFuzzyMatchingU16Start(fanoutFork.BaseOffset, fanoutFork.FuzzyTracker.BestU16Depth, trackSet.Value, subRoutingCode)
		} else {
			result += makeCodeBlockFuzzyMatchingU8U16Middle(trackSet.Value, subRoutingCode)
		}
	}
	return
}

func (inst *CodeGenerateInstance) generateFuzzyMatching(fanoutFork *FanoutFork) (result string) {
	switch fanoutFork.FuzzyModeBit {
	case 8:
		return inst.generateFuzzyMatchingU8(fanoutFork)
	case 16:
		return inst.generateFuzzyMatchingU16(fanoutFork)
	}
	return fmt.Sprintf("// ERROR(generateFuzzyMatching): unknown fuzzy mode bit - %d.", fanoutFork.FuzzyModeBit)
}

func (inst *CodeGenerateInstance) generateInvokeHandler(fanoutFork *FanoutFork) (result string) {
	handlerName := fanoutFork.InvokeHandlerFanout.Route.HandlerName
	result = fmt.Sprintf("%s.%s(w, req, reqPathOffset%s",
		inst.ReceiverName,
		handlerName,
		codeTemplateGenIntPlus(fanoutFork.BaseOffset))
	for _, paramName := range fanoutFork.AvailableSequenceVarName {
		result = result + ", " + paramName
	}
	result += ")\n"
	result += "return " + inst.makeRouteIdentName(handlerName) + ", nil"
	return
}

func (inst *CodeGenerateInstance) generateFanoutCode(fanoutFork *FanoutFork) (result string) {
	switch fanoutFork.LogicType {
	case LogicTypePrefixMatching:
		return inst.generatePrefixMatching(fanoutFork)
	case LogicTypeFuzzyMatching:
		return inst.generateFuzzyMatching(fanoutFork)
	case LogicTypeInvokeHandler:
		return inst.generateInvokeHandler(fanoutFork)
	}
	return fmt.Sprintf("// ERROR: unknown logic type: %v (%v)", fanoutFork.LogicType, fanoutFork.CoveredTerminals)
}

func (inst *CodeGenerateInstance) writeRouteIdentConstants() (err error) {
	codeText := makeCodeTypeRouteIdent(inst.NamePrefix)
	if _, err = inst.fp.WriteString(codeText); nil != err {
		return
	}
	codeText = inst.generateRouteIdentDefinitionListCode()
	_, err = inst.fp.WriteString(codeText)
	return
}

func (inst *CodeGenerateInstance) writePrefixMatchingDigest32Runtime() (routingVarCode string, err error) {
	if !inst.UsePrefixMatching {
		return
	}
	routingVarCode = "var digest32 uint32\n"
	_, err = inst.fp.WriteString(codeFunctionComputePrefixMatching32)
	return
}

// Generate code with given configuration.
func (inst *CodeGenerateInstance) Generate() (err error) {
	if err = inst.validateConfiguration(); nil != err {
		return
	}
	if _, err = inst.fp.WriteString("package " + inst.PackageName + "\n\n"); nil != err {
		return
	}
	if err = inst.writeRouteIdentConstants(); nil != err {
		return
	}
	var routingLogicCode string
	var varDefineCode string
	if varDefineCode, err = inst.writePrefixMatchingDigest32Runtime(); nil != err {
		return
	}
	routingLogicCode += varDefineCode
	routingLogicCode += cleanupCodeBlock(inst.generateFanoutCode(inst.rootFanoutFork), true)
	methodCode := makeCodeMethodRouteEnterance(inst.NamePrefix, routingLogicCode)
	if _, err = inst.fp.WriteString(methodCode); nil != err {
		return
	}
	return nil
}
