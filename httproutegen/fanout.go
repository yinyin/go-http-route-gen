package httproutegen

// FanoutEntry map to a RouteEntry to represent progress of route fanout.
type FanoutEntry struct {
	Route        *RouteEntry    `json:"route"`
	Fanouts      []*FanoutEntry `json:"fanouts,omitempty"`
	Symbols      []Symbol       `json:"symbols,omitempty"`
	CanFork      bool           `json:"can_fork,omitempty"`
	CanTerminate bool           `json:"can_terminate,omitempty"`
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
	return
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
	return
}
