package httproutegen

// FanoutEntry map to a RouteEntry to represent progress of route fanout.
type FanoutEntry struct {
	Route        *RouteEntry    `json:"route"`
	Fanouts      []*FanoutEntry `json:"fanouts,omitempty"`
	CanFork      bool           `json:"can_fork,omitempty"`
	CanTerminate bool           `json:"can_terminate,omitempty"`
}

// NewFanoutEntry maps given RouteEntry and sub-route entries to FanoutEntry.
func NewFanoutEntry(routeEntry *RouteEntry) (fanoutEntry *FanoutEntry) {
	fanoutEntry = &FanoutEntry{
		Route: routeEntry,
	}
	for _, childRoute := range routeEntry.Routes {
		childFanout := NewFanoutEntry(childRoute)
		fanoutEntry.Fanouts = append(fanoutEntry.Fanouts, childFanout)
	}
	return
}

// FanoutInstance expands the route entries
type FanoutInstance struct {
}
