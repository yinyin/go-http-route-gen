package httproutegen

import (
	"strings"
)

// FanoutEntry map to a RouteEntry to represent progress of route fanout.
type FanoutEntry struct {
	Serial            int32          `json:"fanout_serial,omitempty"`
	Route             *RouteEntry    `json:"route"`
	Fanouts           []*FanoutEntry `json:"fanouts,omitempty"`
	Symbols           []Symbol       `json:"symbols,omitempty"`
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

// AssignSerial set serial numbers to given entry and all sub-entries.
func (entry *FanoutEntry) AssignSerial(serialFrom int32) int32 {
	entry.Serial = serialFrom
	serialFrom++
	for _, fo := range entry.Fanouts {
		serialFrom = fo.AssignSerial(serialFrom)
	}
	return serialFrom
}

// FanoutFork track status of an expanding branch of fanout.
type FanoutFork struct {
	CurrentFanouts     []*FanoutEntry
	CurrentSymbolIndex int
}

// FanoutInstance expands the route entries
type FanoutInstance struct {
	InstanceSymbolScope SymbolScope  `json:"symbol_scope"`
	RootFanoutEntry     *FanoutEntry `json:"root_fanout"`
}

// MakeFanoutInstance creates new fanout operation instance from root route entry.
func MakeFanoutInstance(rootRouteEntry *RouteEntry) (instance *FanoutInstance, err error) {
	instance = &FanoutInstance{}
	if instance.RootFanoutEntry, err = MakeFanoutEntry(&instance.InstanceSymbolScope, rootRouteEntry); nil != err {
		return nil, err
	}
	instance.RootFanoutEntry.AssignSerial(1)
	return
}
