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
