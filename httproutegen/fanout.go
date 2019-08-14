package httproutegen

import (
	"fmt"
	"log"
	"strings"
)

// FanoutSymbol is reference to tuple of FanoutEntry and Symbol.
type FanoutSymbol struct {
	Fanout *FanoutEntry
	Symbol *Symbol
}

// CollectAreaNameFromFanoutSymbols get area name from symbols and check if divergent.
func CollectAreaNameFromFanoutSymbols(fanoutSymbols []FanoutSymbol) (areaName string, isDivergent bool) {
	if len(fanoutSymbols) < 1 {
		return
	}
	areaName = fanoutSymbols[0].Fanout.Route.AreaName
	for _, fanoutSymbol := range fanoutSymbols[1:] {
		if fanoutSymbol.Fanout.Route.AreaName != areaName {
			isDivergent = true
		}
	}
	return
}

// FanoutEntry map to a RouteEntry to represent progress of route fanout.
type FanoutEntry struct {
	Serial                 int32          `json:"fanout_serial,omitempty"`
	Route                  *RouteEntry    `json:"route"`
	Fanouts                []*FanoutEntry `json:"fanouts,omitempty"`
	Symbols                []Symbol       `json:"symbols,omitempty"`
	TerminateSerials       []int32        `json:"terminate_fanout_serials,omitempty"`
	MatchSymbolDepthStart  int            `json:"match_symbol_start,omitempty"`
	MatchSymbolDepthFinish int            `json:"match_symbol_finish,omitempty"`
	ParameterCount         int            `json:"parameter_count,omitempty"`
}

// MakeFanoutEntry maps given RouteEntry and sub-route entries to FanoutEntry.
func MakeFanoutEntry(symbolScope *SymbolScope, routeEntry *RouteEntry) (fanoutEntry *FanoutEntry, err error) {
	symbols, err := symbolScope.ParseComponent([]byte(routeEntry.Component))
	if nil != err {
		err = newErrParseComponent(routeEntry.Ident, err)
		return
	}
	if (len(routeEntry.Routes) > 0) && (len(symbols) > 0) {
		symbols = append(symbols, newByteSymbol('/'))
	}
	fanoutEntry = &FanoutEntry{
		Route:   routeEntry,
		Symbols: symbols,
	}
	for _, childRoute := range routeEntry.Routes {
		childFanout, err := MakeFanoutEntry(symbolScope, childRoute)
		if nil != err {
			err = newErrParseComponent(routeEntry.Ident, err)
			return nil, err
		}
		fanoutEntry.Fanouts = append(fanoutEntry.Fanouts, childFanout)
	}
	if err = fanoutEntry.updateMatchSymbolDepth(symbolScope, 0); nil != err {
		return nil, err
	}
	return
}

func (entry *FanoutEntry) updateMatchSymbolDepth(symbolScope *SymbolScope, headingSymbolDepth int) error {
	if entry.Route.StrictMatch {
		entry.MatchSymbolDepthStart = headingSymbolDepth
		entry.MatchSymbolDepthFinish = headingSymbolDepth + len(entry.Symbols) - 1
	} else if ("" != entry.Route.StrictPrefixMatch) &&
		strings.HasPrefix(entry.Route.Component, entry.Route.StrictPrefixMatch) {
		symbols, err := symbolScope.ParseComponent([]byte(entry.Route.StrictPrefixMatch))
		if nil != err {
			return newErrParseComponent(entry.Route.Ident+"::StrictPrefixMatch", err)
		}
		entry.MatchSymbolDepthStart = headingSymbolDepth
		entry.MatchSymbolDepthFinish = headingSymbolDepth + len(symbols) - 1
	} else {
		entry.MatchSymbolDepthStart = -1
		entry.MatchSymbolDepthFinish = -1
	}
	headingSymbolDepth = headingSymbolDepth + len(entry.Symbols)
	for _, fo := range entry.Fanouts {
		if err := fo.updateMatchSymbolDepth(symbolScope, headingSymbolDepth); nil != err {
			return err
		}
	}
	return nil
}

// WithinMatchingDepthRange check if given symbol depth in the range of strict matching.
func (entry *FanoutEntry) WithinMatchingDepthRange(symbolDepth int) bool {
	if (entry.MatchSymbolDepthStart <= symbolDepth) && (entry.MatchSymbolDepthFinish >= symbolDepth) {
		return true
	}
	return false
}

func (entry *FanoutEntry) collectTerminateSerials() (result []int32) {
	if len(entry.Fanouts) == 0 {
		result = append(result, entry.Serial)
	} else {
		for _, fo := range entry.Fanouts {
			aux := fo.collectTerminateSerials()
			result = append(result, aux...)
		}
		entry.TerminateSerials = result
	}
	return
}

func (entry *FanoutEntry) updateParameterCount(parentParameterCount int) {
	currentParameterCount := 0
	for _, sym := range entry.Symbols {
		if sym.Type == SymbolTypeSequence {
			currentParameterCount++
		}
	}
	entry.ParameterCount = parentParameterCount + currentParameterCount
	for _, fo := range entry.Fanouts {
		fo.updateParameterCount(entry.ParameterCount)
	}
}

// GetTerminateSerials return terminate serials in slice for this fanout entry.
func (entry *FanoutEntry) GetTerminateSerials() (result []int32) {
	if len(entry.Fanouts) == 0 {
		result = []int32{entry.Serial}
	} else {
		result = entry.TerminateSerials
	}
	return
}

// AssignSerial set serial numbers to given entry and all sub-entries.
func (entry *FanoutEntry) AssignSerial(serialFrom int32) int32 {
	entry.Serial = serialFrom
	serialFrom++
	for _, fo := range entry.Fanouts {
		serialFrom = fo.AssignSerial(serialFrom)
	}
	return serialFrom
}

// GetSymbol return symbols in given depth.
func (entry *FanoutEntry) GetSymbol(depth int) (result []FanoutSymbol) {
	if depth < 0 {
		log.Printf("WARN: request symbol at %d which cannot be reach.", depth)
		return
	}
	if len(entry.Symbols) > depth {
		aux := FanoutSymbol{
			Fanout: entry,
			Symbol: &entry.Symbols[depth],
		}
		result = append(result, aux)
		return
	}
	subDepth := depth - len(entry.Symbols)
	for _, fo := range entry.Fanouts {
		syms := fo.GetSymbol(subDepth)
		if len(syms) > 0 {
			result = append(result, syms...)
		}
	}
	return
}

// FindFanoutEntryBySerial search for given serial at current entry and child entries.
func (entry *FanoutEntry) FindFanoutEntryBySerial(serial int32) *FanoutEntry {
	if entry.Serial == serial {
		return entry
	}
	for _, fo := range entry.Fanouts {
		if result := fo.FindFanoutEntryBySerial(serial); nil != result {
			return result
		}
	}
	return nil
}

func isTerminateSerialsCoveredFanoutSymbol(terminateSerials []int32, symbol FanoutSymbol) bool {
	for _, serial := range terminateSerials {
		if serial == symbol.Fanout.Serial {
			return true
		}
		for _, ts := range symbol.Fanout.TerminateSerials {
			if ts == serial {
				return true
			}
		}
	}
	return false
}

// FanoutLiteralDigestSet is a group of fanouts share same literal digest value.
type FanoutLiteralDigestSet struct {
	TerminateSerials []int32
	Value            uint32
}

// Covered check if given symbol is covered in this digest set
func (s *FanoutLiteralDigestSet) Covered(symbol FanoutSymbol) bool {
	return isTerminateSerialsCoveredFanoutSymbol(s.TerminateSerials, symbol)
}

// FanoutLiteralDigestPartition is a group of FanoutLiteralDigestSet
type FanoutLiteralDigestPartition struct {
	Digests []*FanoutLiteralDigestSet
	Depth   int
}

func (p *FanoutLiteralDigestPartition) searchDigestSet(symbol FanoutSymbol) int {
	for idx, dst := range p.Digests {
		if dst.Covered(symbol) {
			return idx
		}
	}
	return -1
}

// FeedSymbols save symbols into digest sets
func (p *FanoutLiteralDigestPartition) FeedSymbols(symbols []FanoutSymbol) {
	var updatedSet []*FanoutLiteralDigestSet
	for _, sym := range symbols {
		var digestValue uint32
		if dstIdx := p.searchDigestSet(sym); dstIdx < 0 {
			digestValue = uint32(sym.Symbol.ByteValue)
		} else {
			digestValue = (p.Digests[dstIdx].Value << 8) | uint32(sym.Symbol.ByteValue)
		}
		attached := false
		for _, s := range updatedSet {
			if s.Value == digestValue {
				s.TerminateSerials = append(s.TerminateSerials, sym.Fanout.GetTerminateSerials()...)
				attached = true
				break
			}
		}
		if !attached {
			aux := FanoutLiteralDigestSet{
				Value: digestValue,
			}
			aux.TerminateSerials = append(aux.TerminateSerials, sym.Fanout.GetTerminateSerials()...)
			updatedSet = append(updatedSet, &aux)
		}
	}
	p.Depth++
	p.Digests = updatedSet
}

// CoveredTerminalCount compute number of covered terminal serials.
func (p *FanoutLiteralDigestPartition) CoveredTerminalCount() (totalCoveredTerminals int) {
	for _, digestSet := range p.Digests {
		totalCoveredTerminals += len(digestSet.TerminateSerials)
	}
	return
}

// FanoutFuzzyTrackSet is collect of fanout share same value for fuzzy tracking.
type FanoutFuzzyTrackSet struct {
	TerminateSerials []int32
	FanoutSymbols    []*FanoutSymbol
	Value            uint32
}

// Covered check if symbol is covered by this set.
func (s *FanoutFuzzyTrackSet) Covered(symbol FanoutSymbol) bool {
	return isTerminateSerialsCoveredFanoutSymbol(s.TerminateSerials, symbol)
}

func appendFanoutFuzzyTrackSet(targetSet []*FanoutFuzzyTrackSet, value uint32, sym FanoutSymbol) []*FanoutFuzzyTrackSet {
	attached := false
	for _, s := range targetSet {
		if s.Value == value {
			s.TerminateSerials = append(s.TerminateSerials, sym.Fanout.GetTerminateSerials()...)
			s.FanoutSymbols = append(s.FanoutSymbols, &sym)
			attached = true
			break
		}
	}
	if !attached {
		aux := FanoutFuzzyTrackSet{
			Value: value,
		}
		aux.TerminateSerials = append(aux.TerminateSerials, sym.Fanout.GetTerminateSerials()...)
		aux.FanoutSymbols = append(aux.FanoutSymbols, &sym)
		targetSet = append(targetSet, &aux)
	}
	return targetSet
}

// FanoutFuzzyTrackPartition is collect of FanoutFuzzyTrackSet.
type FanoutFuzzyTrackPartition struct {
	FrontU16 []*FanoutFuzzyTrackSet

	BestU8       []*FanoutFuzzyTrackSet
	BestU8Depth  int
	BestU16      []*FanoutFuzzyTrackSet
	BestU16Depth int

	Depth int
}

func (p *FanoutFuzzyTrackPartition) searchFrontU16TrackSet(symbol FanoutSymbol) int {
	for idx, dst := range p.FrontU16 {
		if dst.Covered(symbol) {
			return idx
		}
	}
	return -1
}

// FeedSymbols save symbols into fuzzy track sets
func (p *FanoutFuzzyTrackPartition) FeedSymbols(symbols []FanoutSymbol) {
	var updatedU8Set []*FanoutFuzzyTrackSet
	var updatedU16Set []*FanoutFuzzyTrackSet
	for _, sym := range symbols {
		var digestValueU8 = uint32(sym.Symbol.ByteValue)
		updatedU8Set = appendFanoutFuzzyTrackSet(updatedU8Set, digestValueU8, sym)
		var digestValueU16 uint32
		if dstIdx := p.searchFrontU16TrackSet(sym); dstIdx < 0 {
			digestValueU16 = uint32(sym.Symbol.ByteValue)
		} else {
			digestValueU16 = ((p.FrontU16[dstIdx].Value << 8) | uint32(sym.Symbol.ByteValue)) & 0xFFFF
		}
		updatedU16Set = appendFanoutFuzzyTrackSet(updatedU16Set, digestValueU16, sym)
	}
	if len(updatedU8Set) > len(p.BestU8) {
		p.BestU8 = updatedU8Set
		p.BestU8Depth = p.Depth
	}
	if len(updatedU16Set) > len(p.BestU16) {
		p.BestU16 = updatedU16Set
		p.BestU16Depth = p.Depth
	}
	p.FrontU16 = updatedU16Set
	p.Depth++
}

// CoveredTerminalCount compute number of covered terminal serials.
func (p *FanoutFuzzyTrackPartition) CoveredTerminalCount() (totalCoveredTerminals int) {
	for _, trackSet := range p.BestU8 {
		totalCoveredTerminals += len(trackSet.TerminateSerials)
	}
	return
}

// FanoutForkLogicType is the logic type of a fanout branch.
type FanoutForkLogicType int

// Logic types of an fanout fork branching node.
const (
	LogicTypeUnknown FanoutForkLogicType = iota
	LogicTypePrefixMatching
	LogicTypeFuzzyMatching
	LogicTypeGetParameter
	LogicTypeInvokeHandler
)

// FanoutFork track status of an expanding branch of fanout.
type FanoutFork struct {
	CoveredTerminals []int32             `json:"convered_terminals"`
	ParentFork       *FanoutFork         `json:"-"`
	ChildForks       []*FanoutFork       `json:"child_forks,omitempty"`
	AreaName         string              `json:"area,omitempty"`
	LogicType        FanoutForkLogicType `json:"logic_type"`

	BaseOffset int `json:"base_offset"`

	MaxMatchingDepth     int `json:"max_matching_depth,omitempty"`
	PrefixLiteralDigests FanoutLiteralDigestPartition

	FuzzyTracker FanoutFuzzyTrackPartition
	FuzzyModeBit int `json:"fuzzy_mode_bit,omitempty"`

	SequenceIndex            int      `json:"sequence_index,omitempty"`
	SequenceVarName          string   `json:"sequence_variable,omitempty"`
	AvailableSequenceVarName []string `json:"available_sequence_variable,omitempty"`

	InvokeHandlerFanout *FanoutEntry `json:"invoke_handler,omitempty"`
}

// Covered check if given fanout-symbol is covered by this fork.
func (fork *FanoutFork) Covered(fanoutSymbol FanoutSymbol) bool {
	return isTerminateSerialsCoveredFanoutSymbol(fork.CoveredTerminals, fanoutSymbol)
}

// FullyMatch check if covered terminals are fully equal to given fanout symbol.
func (fork *FanoutFork) FullyMatch(fanoutSymbol FanoutSymbol) bool {
	termSerials := fanoutSymbol.Fanout.GetTerminateSerials()
	if len(fork.CoveredTerminals) != len(termSerials) {
		log.Printf("INFO: terminate serial not fully match: %v vs. %v", fork.CoveredTerminals, termSerials)
		return false
	}
	for _, serial := range fork.CoveredTerminals {
		found := false
		for _, ts := range termSerials {
			if serial == ts {
				found = true
				break
			}
		}
		if !found {
			log.Printf("INFO: terminate serial not fully match: %v vs. %v", fork.CoveredTerminals, termSerials)
			return false
		}
	}
	return true
}

func (fork *FanoutFork) haveCoveredTerminateSerials(terminalSerial []int32) bool {
	for _, ts := range fork.CoveredTerminals {
		for _, oth := range terminalSerial {
			if ts == oth {
				return true
			}
		}
	}
	return false
}

// FindChildForkViaTerminateSerials search for child forks covered by given terminate serials.
func (fork *FanoutFork) FindChildForkViaTerminateSerials(terminalSerial []int32) (result []*FanoutFork) {
	for _, childFork := range fork.ChildForks {
		if childFork.haveCoveredTerminateSerials(terminalSerial) {
			result = append(result, childFork)
		}
	}
	return
}

func (fork *FanoutFork) chooseLogicType(symbols []FanoutSymbol, symbolDepth int) FanoutForkLogicType {
	maxMatchingDepth := 0
	symbolType := SymbolTypeNoop
	for _, sym := range symbols {
		if sym.Symbol.Type != symbolType {
			if symbolType == SymbolTypeNoop {
				symbolType = sym.Symbol.Type
			} else {
				log.Fatalf("mixed symbol type: %#v", symbols)
			}
		}
		if (maxMatchingDepth < sym.Fanout.MatchSymbolDepthFinish) && sym.Fanout.WithinMatchingDepthRange(symbolDepth) {
			maxMatchingDepth = sym.Fanout.MatchSymbolDepthFinish
		}
	}
	switch symbolType {
	case SymbolTypeNoop:
		log.Fatalf("not reaching usable logic type: %#v", symbols)
	case SymbolTypeByte:
		if maxMatchingDepth > 0 {
			fork.MaxMatchingDepth = maxMatchingDepth
			return LogicTypePrefixMatching
		}
	case SymbolTypeSequence:
		return LogicTypeGetParameter
	}
	if len(symbols) == 1 {
		return LogicTypeUnknown
	}
	return LogicTypeFuzzyMatching
}

func (fork *FanoutFork) makeNextStageForksFromPrefixMatching() (nextStageForks []*FanoutFork) {
	for _, s := range fork.PrefixLiteralDigests.Digests {
		aux := FanoutFork{
			BaseOffset: 0, // fork.BaseOffset + fork.PrefixLiteralDigests.Depth,
			AreaName:   fork.AreaName,
		}
		aux.CoveredTerminals = append(aux.CoveredTerminals, s.TerminateSerials...)
		aux.AvailableSequenceVarName = append(aux.AvailableSequenceVarName, fork.AvailableSequenceVarName...)
		nextStageForks = append(nextStageForks, &aux)
	}
	return
}

func (fork *FanoutFork) rejectSymbolWithSealPrefixMatching(symbols []FanoutSymbol) (reject bool, nextStageForks []*FanoutFork, err error) {
	reject = true
	nextStageForks = fork.makeNextStageForksFromPrefixMatching()
	for _, sym := range symbols {
		switch sym.Symbol.Type {
		case SymbolTypeSequence:
			fo := FindFanoutForkForSymbol(nextStageForks, sym)
			if !fo.FullyMatch(sym) {
				err = fmt.Errorf("parameter fetch is not fully matching fork: %#v, %v", sym.Fanout, fo.LogicType)
			}
		}
	}
	return
}

func (fork *FanoutFork) feedSymbolsToPrefixMatching(symbols []FanoutSymbol, symbolDepth int) (reject bool, nextStageForks []*FanoutFork, err error) {
	totalCoveredTerminals := 0
	for _, sym := range symbols {
		if sym.Symbol.Type != SymbolTypeByte {
			return fork.rejectSymbolWithSealPrefixMatching(symbols)
		}
		totalCoveredTerminals += len(sym.Fanout.GetTerminateSerials())
	}
	if totalCoveredTerminals < fork.PrefixLiteralDigests.CoveredTerminalCount() {
		log.Printf("INFO: feedSymbolsToPrefixMatching - covered terminal count shrink: %d <- %d", totalCoveredTerminals, fork.PrefixLiteralDigests.CoveredTerminalCount())
		return fork.rejectSymbolWithSealPrefixMatching(symbols)
	}
	fork.PrefixLiteralDigests.FeedSymbols(symbols)
	if (fork.MaxMatchingDepth == symbolDepth) || (fork.PrefixLiteralDigests.Depth == 4) {
		return false, fork.makeNextStageForksFromPrefixMatching(), nil
	}
	return false, nil, nil
}

func (fork *FanoutFork) makeNextStageForksFromFuzzyMatching() (nextStageForks []*FanoutFork) {
	if fork.FuzzyModeBit == 0 {
		if len(fork.FuzzyTracker.BestU16) > len(fork.FuzzyTracker.BestU8) {
			fork.FuzzyModeBit = 16
		} else {
			fork.FuzzyModeBit = 8
		}
	}
	var trackSet []*FanoutFuzzyTrackSet
	var trackDepth int
	if fork.FuzzyModeBit == 16 {
		trackSet = fork.FuzzyTracker.BestU16
		trackDepth = fork.FuzzyTracker.BestU16Depth
	} else {
		trackSet = fork.FuzzyTracker.BestU8
		trackDepth = fork.FuzzyTracker.BestU8Depth
	}
	for _, s := range trackSet {
		aux := FanoutFork{
			BaseOffset: fork.FuzzyTracker.Depth - trackDepth, // fork.BaseOffset + fork.FuzzyTracker.Depth,
			AreaName:   fork.AreaName,
		}
		aux.CoveredTerminals = append(aux.CoveredTerminals, s.TerminateSerials...)
		aux.AvailableSequenceVarName = append(aux.AvailableSequenceVarName, fork.AvailableSequenceVarName...)
		nextStageForks = append(nextStageForks, &aux)
	}
	return
}

func (fork *FanoutFork) rejectSymbolWithSealFuzzyMatching(symbols []FanoutSymbol) (reject bool, nextStageForks []*FanoutFork, err error) {
	reject = true
	nextStageForks = fork.makeNextStageForksFromFuzzyMatching()
	for _, sym := range symbols {
		switch sym.Symbol.Type {
		case SymbolTypeSequence:
			fo := FindFanoutForkForSymbol(nextStageForks, sym)
			if !fo.FullyMatch(sym) {
				err = fmt.Errorf("parameter fetch is not fully matching fork: %#v, %v", sym.Fanout, fo.LogicType)
			}
		}
	}
	return
}

func (fork *FanoutFork) feedSymbolsToFuzzyMatching(symbols []FanoutSymbol, symbolDepth int) (reject bool, nextStageForks []*FanoutFork, err error) {
	totalCoveredTerminals := 0
	for _, sym := range symbols {
		if (sym.Symbol.Type != SymbolTypeByte) || sym.Fanout.WithinMatchingDepthRange(symbolDepth) {
			return fork.rejectSymbolWithSealFuzzyMatching(symbols)
		}
		totalCoveredTerminals += len(sym.Fanout.GetTerminateSerials())
	}
	if totalCoveredTerminals < fork.FuzzyTracker.CoveredTerminalCount() {
		log.Printf("INFO: feedSymbolsToFuzzyMatching - covered terminal count shrink: %d <- %d", totalCoveredTerminals, fork.FuzzyTracker.CoveredTerminalCount())
		return fork.rejectSymbolWithSealFuzzyMatching(symbols)
	}
	fork.FuzzyTracker.FeedSymbols(symbols)
	return false, nil, nil
}

func (fork *FanoutFork) makeNextStageForksFromGetParameter() (nextStageForks []*FanoutFork) {
	aux := FanoutFork{
		CoveredTerminals: fork.CoveredTerminals,
		AreaName:         fork.AreaName,
	}
	aux.AvailableSequenceVarName = append(aux.AvailableSequenceVarName, fork.AvailableSequenceVarName...)
	nextStageForks = append(nextStageForks, &aux)
	return
}

func (fork *FanoutFork) feedSymbolsToGetParameter(symbols []FanoutSymbol) (reject bool, nextStageForks []*FanoutFork, err error) {
	for _, sym := range symbols {
		if sym.Symbol.Type != SymbolTypeSequence {
			log.Fatalf("should be sequence symbol: %#v", sym.Symbol)
		}
		if fork.SequenceVarName == "" {
			for _, existedVarName := range fork.AvailableSequenceVarName {
				if existedVarName == sym.Symbol.SequenceVarName {
					return true, nil, fmt.Errorf("sequence variable name existed: %v, %v", fork.AvailableSequenceVarName, sym.Symbol.SequenceVarName)
				}
			}
			fork.SequenceIndex = sym.Symbol.SequenceIndex
			fork.SequenceVarName = sym.Symbol.SequenceVarName
			fork.AvailableSequenceVarName = append(fork.AvailableSequenceVarName, sym.Symbol.SequenceVarName)
		} else if (fork.SequenceIndex != sym.Symbol.SequenceIndex) || (fork.SequenceVarName != sym.Symbol.SequenceVarName) {
			return true, nil, fmt.Errorf("incompatible sequence: (%v, %v) != %#v", fork.SequenceIndex, fork.SequenceVarName, sym)
		}
	}
	if fork.SequenceVarName == "" {
		return true, nil, fmt.Errorf("empty sequence variable name: %#v", symbols)
	}
	return false, fork.makeNextStageForksFromGetParameter(), nil
}

// FeedSymbols get symbols from fanouts and update logic state for code generation.
func (fork *FanoutFork) FeedSymbols(symbols []FanoutSymbol, symbolDepth int) (reject bool, nextStageForks []*FanoutFork, err error) {
	if LogicTypeUnknown == fork.LogicType {
		fork.LogicType = fork.chooseLogicType(symbols, symbolDepth)
	}
	switch fork.LogicType {
	case LogicTypePrefixMatching:
		return fork.feedSymbolsToPrefixMatching(symbols, symbolDepth)
	case LogicTypeFuzzyMatching:
		return fork.feedSymbolsToFuzzyMatching(symbols, symbolDepth)
	case LogicTypeGetParameter:
		return fork.feedSymbolsToGetParameter(symbols)
	case LogicTypeUnknown:
		fork.BaseOffset++
		return false, nil, nil
	}
	return true, nil, fmt.Errorf("unknown logic type: %v", fork.LogicType)
}

func (fork *FanoutFork) divideThisFork() (hasDivide bool) {
	if fork.LogicType == LogicTypeFuzzyMatching {
		nextStageForks := fork.makeNextStageForksFromFuzzyMatching()
		fork.ChildForks = nextStageForks
		for _, childFork := range nextStageForks {
			childFork.ParentFork = fork
		}
		return true
	}
	return false
}

func (fork *FanoutFork) sealThisFork(rootFanoutEntry *FanoutEntry) (stopPropagate bool) {
	if len(fork.CoveredTerminals) != 1 {
		if !fork.divideThisFork() {
			log.Fatalf("ERROR: does not terminate with one and only one terminate serial: %#v", fork)
		}
		return false
	}
	handlerFanout := rootFanoutEntry.FindFanoutEntryBySerial(fork.CoveredTerminals[0])
	if fork.LogicType == LogicTypeUnknown {
		fork.LogicType = LogicTypeInvokeHandler
		fork.InvokeHandlerFanout = handlerFanout
	} else {
		aux := FanoutFork{
			AreaName:            fork.AreaName,
			LogicType:           LogicTypeInvokeHandler,
			CoveredTerminals:    fork.CoveredTerminals,
			ParentFork:          fork,
			InvokeHandlerFanout: handlerFanout,
		}
		aux.AvailableSequenceVarName = append(aux.AvailableSequenceVarName, fork.AvailableSequenceVarName...)
		fork.ChildForks = []*FanoutFork{&aux}
	}
	return true
}

// SealTerminateFork mark or create invoke fork for terminate fork.
func (fork *FanoutFork) SealTerminateFork(rootFanoutEntry *FanoutEntry) {
	if len(fork.ChildForks) == 0 {
		if fork.sealThisFork(rootFanoutEntry) {
			return
		}
	}
	for _, childFork := range fork.ChildForks {
		childFork.SealTerminateFork(rootFanoutEntry)
	}
}

// FindFanoutForkForSymbol search for FanoutFork via symbol coverage.
func FindFanoutForkForSymbol(forks []*FanoutFork, symbol FanoutSymbol) *FanoutFork {
	for _, fo := range forks {
		if fo.Covered(symbol) {
			return fo
		}
	}
	log.Fatalf("ERROR: cannot reach fork for symbol: %#v, from %v", symbol, forks)
	return nil
}

// FanoutForkSlice package operations for series of FanoutForks
type FanoutForkSlice struct {
	Forks []*FanoutFork
}

// AttachParentFork attach given fork to elements of slice as parent fork.
func (s *FanoutForkSlice) AttachParentFork(fork *FanoutFork) {
	for _, fo := range s.Forks {
		if fo == fork {
			continue
		}
		if fo.ParentFork == fork {
			continue
		}
		if fo.ParentFork != nil {
			log.Printf("parent fork existed: %v", fo)
		}
		fo.ParentFork = fork
		fork.ChildForks = append(fork.ChildForks, fo)
	}
}

// UpdateAreaName set area name of fork to given areaName.
func (s *FanoutForkSlice) UpdateAreaName(areaName string) {
	for _, fo := range s.Forks {
		fo.AreaName = areaName
	}
}

func (s *FanoutForkSlice) distributeSymbols(symbols []FanoutSymbol) [][]FanoutSymbol {
	symbolBuckets := make([][]FanoutSymbol, len(s.Forks))
	for _, sym := range symbols {
		emitted := false
		for idx, fanout := range s.Forks {
			if !fanout.Covered(sym) {
				continue
			}
			symbolBuckets[idx] = append(symbolBuckets[idx], sym)
			emitted = true
			break
		}
		if !emitted {
			log.Fatalf("ERROR: symbol failed to emit into fork bucket: %#v", sym)
		}
	}
	return symbolBuckets
}

// FeedSymbols feed symbols into covered FanoutFork.
// The slice will be update if fork is forked further.
func (s *FanoutForkSlice) FeedSymbols(symbols []FanoutSymbol, symbolDepth int) error {
	symbolBuckets := s.distributeSymbols(symbols)
	var updatedForks []*FanoutFork
	for idx, fanout := range s.Forks {
		if len(symbolBuckets[idx]) == 0 {
			updatedForks = append(updatedForks, fanout)
			continue
		}
		areaName, isDivergent := CollectAreaNameFromFanoutSymbols(symbolBuckets[idx])
		if reject, nextStageForks, err := fanout.FeedSymbols(symbolBuckets[idx], symbolDepth); nil != err {
			return err
		} else if len(nextStageForks) > 0 {
			subSlice := FanoutForkSlice{
				Forks: nextStageForks,
			}
			subSlice.AttachParentFork(fanout)
			if !isDivergent {
				subSlice.UpdateAreaName(areaName)
			}
			if reject {
				if err = subSlice.FeedSymbols(symbolBuckets[idx], symbolDepth); nil != err {
					return err
				}
			}
			updatedForks = append(updatedForks, subSlice.Forks...)
		} else {
			if reject {
				log.Fatalf("ERROR: rejected FeedSymbols() must return new forks: %#v; [PARENT]: %#v", fanout, fanout.ParentFork)
			}
			updatedForks = append(updatedForks, fanout)
		}
	}
	s.Forks = updatedForks
	return nil
}

// FanoutInstance expands the route entries
type FanoutInstance struct {
	InstanceSymbolScope SymbolScope  `json:"symbol_scope"`
	RootFanoutEntry     *FanoutEntry `json:"root_fanout"`

	RootFanoutFork *FanoutFork `json:"root_fork"`
}

// MakeFanoutInstance creates new fanout operation instance from root route entry.
func MakeFanoutInstance(rootRouteEntry *RouteEntry) (instance *FanoutInstance, err error) {
	instance = &FanoutInstance{}
	if instance.RootFanoutEntry, err = MakeFanoutEntry(&instance.InstanceSymbolScope, rootRouteEntry); nil != err {
		return nil, err
	}
	instance.RootFanoutEntry.AssignSerial(1)
	instance.RootFanoutEntry.collectTerminateSerials()
	instance.RootFanoutEntry.updateParameterCount(0)
	return
}

// ExpandFanout expand fanout entries into fanout fork.
func (instance *FanoutInstance) ExpandFanout() (err error) {
	rootFanoutFork := &FanoutFork{
		CoveredTerminals: instance.RootFanoutEntry.TerminateSerials,
		AreaName:         instance.RootFanoutEntry.Route.AreaName,
	}
	fanoutForks := FanoutForkSlice{
		Forks: []*FanoutFork{rootFanoutFork},
	}
	depth := 0
	symbols := instance.RootFanoutEntry.GetSymbol(depth)
	for len(symbols) > 0 {
		if err = fanoutForks.FeedSymbols(symbols, depth); nil != err {
			return
		}
		depth++
		symbols = instance.RootFanoutEntry.GetSymbol(depth)
	}
	rootFanoutFork.SealTerminateFork(instance.RootFanoutEntry)
	instance.RootFanoutFork = rootFanoutFork
	return nil
}
