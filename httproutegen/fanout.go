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

// FanoutEntry map to a RouteEntry to represent progress of route fanout.
type FanoutEntry struct {
	Serial            int32          `json:"fanout_serial,omitempty"`
	Route             *RouteEntry    `json:"route"`
	Fanouts           []*FanoutEntry `json:"fanouts,omitempty"`
	Symbols           []Symbol       `json:"symbols,omitempty"`
	TerminateSerials  []int32        `json:"terminate_fanout_serials,omitempty"`
	MinForkIndex      int            `json:"min_fork_at,omitempty"`
	MinTerminateIndex int            `json:"min_terminate_at,omitempty"`
}

// MakeFanoutEntry maps given RouteEntry and sub-route entries to FanoutEntry.
func MakeFanoutEntry(symbolScope *SymbolScope, routeEntry *RouteEntry) (fanoutEntry *FanoutEntry, err error) {
	symbols, err := symbolScope.ParseComponent([]byte(routeEntry.Component))
	if nil != err {
		err = newErrParseComponent(routeEntry.Ident, err)
		return
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
	if err = fanoutEntry.updateForkIndex(symbolScope); nil != err {
		return nil, err
	}
	return
}

func (entry *FanoutEntry) updateForkIndex(symbolScope *SymbolScope) error {
	if entry.Route.StrictMatch {
		boundIndex := len(entry.Symbols) - 1
		entry.MinForkIndex = boundIndex
		entry.MinTerminateIndex = boundIndex
		return nil
	}
	if ("" != entry.Route.StrictPrefixMatch) &&
		strings.HasPrefix(entry.Route.Component, entry.Route.StrictPrefixMatch) {
		symbols, err := symbolScope.ParseComponent([]byte(entry.Route.StrictPrefixMatch))
		if nil != err {
			return newErrParseComponent(entry.Route.Ident+"::StrictPrefixMatch", err)
		}
		boundIndex := len(symbols) - 1
		if boundIndex > entry.MinForkIndex {
			entry.MinForkIndex = boundIndex
		}
	}
	minTermIndex := 0
	for idx, sym := range entry.Symbols {
		if sym.Type == SymbolTypeSequence {
			minTermIndex = idx
		}
	}
	if minTermIndex > entry.MinTerminateIndex {
		entry.MinTerminateIndex = minTermIndex
	}
	return entry.updateFanoutForkIndex(symbolScope)
}

func (entry *FanoutEntry) updateFanoutForkIndex(symbolScope *SymbolScope) error {
	for _, fanout := range entry.Fanouts {
		if "" != fanout.Route.StrictPrefixMatch {
			symbols, err := symbolScope.ParseComponent([]byte(fanout.Route.StrictPrefixMatch))
			if nil != err {
				return newErrParseComponent(entry.Route.Ident+"::StrictPrefixMatch(updateFanoutForkIndex)", err)
			}
			boundIndex := len(symbols) - 1
			for _, fo := range entry.Fanouts {
				if strings.HasPrefix(fo.Route.Component, fanout.Route.StrictPrefixMatch) &&
					(fo.MinForkIndex < boundIndex) {
					fo.MinForkIndex = boundIndex
				}
			}
		}
	}
	for _, fanout := range entry.Fanouts {
		fanout.updateForkIndex(symbolScope)
	}
	return nil
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
	if len(entry.Symbols) < depth {
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

// FanoutForkLogicType is the logic type of a fanout branch.
type FanoutForkLogicType int

// Logic types of an fanout fork branching node.
const (
	LogicTypeUnknown FanoutForkLogicType = iota
	LogicTypePrefixMatching
	LogicTypeFuzzyMatching
)

// FanoutFork track status of an expanding branch of fanout.
type FanoutFork struct {
	CoveredTerminals []int32
	ParentFork       *FanoutFork
	ChildForks       []*FanoutFork
	LogicType        FanoutForkLogicType

	MaxPrefixMatchLength int
}

// Covered check if given fanout-symbol is covered by this fork.
func (fork *FanoutFork) Covered(fanoutSymbol FanoutSymbol) bool {
	for _, serial := range fork.CoveredTerminals {
		if serial == fanoutSymbol.Fanout.Serial {
			return true
		}
		for _, ts := range fanoutSymbol.Fanout.TerminateSerials {
			if serial == ts {
				return true
			}
		}
	}
	return false
}

func (fork *FanoutFork) chooseLogicType(symbols []FanoutSymbol) FanoutForkLogicType {
	maxPrefixMatchLength := 0
	for _, sym := range symbols {
		if sym.Fanout.MinForkIndex > maxPrefixMatchLength {
			maxPrefixMatchLength = sym.Fanout.MinForkIndex
		}
	}
	if maxPrefixMatchLength > 0 {
		fork.MaxPrefixMatchLength = maxPrefixMatchLength
		return LogicTypePrefixMatching
	}
	return LogicTypeFuzzyMatching
}

func (fork *FanoutFork) feedSymbolsToPrefixMatching(symbols []FanoutSymbol) (reject bool, nextStageForks []*FanoutFork, err error) {
	// TODO: implement
	return true, nil, nil
}

func (fork *FanoutFork) feedSymbolsToFuzzyMatching(symbols []FanoutSymbol) (reject bool, nextStageForks []*FanoutFork, err error) {
	// TODO: implement
	return true, nil, nil
}

// FeedSymbols get symbols from fanouts and update logic state for code generation.
func (fork *FanoutFork) FeedSymbols(symbols []FanoutSymbol) (reject bool, nextStageForks []*FanoutFork, err error) {
	if LogicTypeUnknown == fork.LogicType {
		fork.LogicType = fork.chooseLogicType(symbols)
	}
	switch fork.LogicType {
	case LogicTypePrefixMatching:
		return fork.feedSymbolsToPrefixMatching(symbols)
	case LogicTypeFuzzyMatching:
		return fork.feedSymbolsToFuzzyMatching(symbols)
	}
	return true, nil, fmt.Errorf("unknown logic type: %v", fork.LogicType)
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
		} else if fo.ParentFork != nil {
			log.Printf("WARN: attaching fork as parent fork to attached fork: %#v <= %#v", fo, fork)
			continue
		}
		fo.ParentFork = fork
		fork.ChildForks = append(fork.ChildForks, fo)
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
func (s *FanoutForkSlice) FeedSymbols(symbols []FanoutSymbol) error {
	symbolBuckets := s.distributeSymbols(symbols)
	var updatedForks []*FanoutFork
	for idx, fanout := range s.Forks {
		if reject, nextStageForks, err := fanout.FeedSymbols(symbolBuckets[idx]); nil != err {
			return err
		} else if len(nextStageForks) > 0 {
			subSlice := FanoutForkSlice{
				Forks: nextStageForks,
			}
			if reject {
				if err = subSlice.FeedSymbols(symbolBuckets[idx]); nil != err {
					return err
				}
			}
			subSlice.AttachParentFork(fanout)
			updatedForks = append(updatedForks, subSlice.Forks...)
		} else {
			if reject {
				log.Fatalf("ERROR: rejected FeedSymbols() must return new forks: %#v", fanout)
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
	return
}

// ExpandFanout expand fanout entries into fanout fork.
func (instance *FanoutInstance) ExpandFanout() (err error) {
	rootFanoutFork := &FanoutFork{
		CoveredTerminals: instance.RootFanoutEntry.TerminateSerials,
	}
	fanoutForks := FanoutForkSlice{
		Forks: []*FanoutFork{rootFanoutFork},
	}
	depth := 0
	symbols := instance.RootFanoutEntry.GetSymbol(depth)
	for len(symbols) > 0 {
		if err = fanoutForks.FeedSymbols(symbols); nil != err {
			return
		}
		depth++
		symbols = instance.RootFanoutEntry.GetSymbol(depth)
	}
	instance.RootFanoutFork = rootFanoutFork
	return nil
}
