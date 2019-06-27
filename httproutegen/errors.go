package httproutegen

// ErrConflictConfiguration represent conflict in configuration
type ErrConflictConfiguration struct {
	Component string
	Config1   string
	Config2   string
	Message   string
}

func (e *ErrConflictConfiguration) Error() string {
	leadingText := "ErrConflictConfiguration: option \"" + e.Config1 + "\" and \"" + e.Config2 + "\" for " + e.Component
	if "" == e.Message {
		return leadingText + " have conflict values"
	}
	return leadingText + " have conflict: " + e.Message
}

// ErrParseComponent represent failure in component parsing.
type ErrParseComponent struct {
	Component string
	Err       error
}

func newErrParseComponent(componentIdent string, err error) error {
	return &ErrParseComponent{
		Component: componentIdent,
		Err:       err,
	}
}

func (e *ErrParseComponent) Error() string {
	return "ErrParseComponent: component=" + e.Component + ", error=" + e.Err.Error()
}
