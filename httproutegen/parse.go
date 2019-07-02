package httproutegen

import (
	"errors"
	"io/ioutil"
	"strings"
	"unicode"

	yaml "gopkg.in/yaml.v2"
)

func shouldTrimFromComponent(ch rune) bool {
	return unicode.IsSpace(ch) || (ch == '/')
}

// RouteEntry represent an entry of route
type RouteEntry struct {
	Ident             string        `yaml:"-" json:"component_ident,omitempty"`
	Component         string        `yaml:"c,omitempty" json:"c,omitempty"`
	HandlerName       string        `yaml:"handler,omitempty" json:"handler,omitempty"`
	StrictPrefixMatch string        `yaml:"strict-prefix-match,omitempty" json:"strict_prefix_match,omitempty"`
	StrictMatch       bool          `yaml:"strict-match,omitempty" json:"strict_match,omitempty"`
	TrailingSlash     bool          `yaml:"trailing-slash,omitempty" json:"trailing_slash,omitempty"`
	Routes            []*RouteEntry `yaml:"route,omitempty" json:"route,omitempty"`
}

func (entry *RouteEntry) makeComponentIdent(parentComponentIdent string) string {
	if "" == entry.Component {
		if "" == parentComponentIdent {
			return "/"
		}
		return parentComponentIdent
	}
	return parentComponentIdent + entry.Component + "/"
}

func (entry *RouteEntry) cleanupComponent(parentComponentIdent string) error {
	entry.Component = strings.TrimFunc(entry.Component, shouldTrimFromComponent)
	if ("" == entry.Component) && ("" != parentComponentIdent) {
		return errors.New("empty component: parent=[" + parentComponentIdent + "], handler=[" + entry.HandlerName + "]")
	}
	return nil
}

func (entry *RouteEntry) cleanupStrictPrefixMatch() {
	entry.StrictPrefixMatch = strings.TrimLeftFunc(entry.StrictPrefixMatch, shouldTrimFromComponent)
}

func (entry *RouteEntry) verifyConfiguration(parentComponentIdent string) error {
	if err := entry.cleanupComponent(parentComponentIdent); nil != err {
		return err
	}
	entry.cleanupStrictPrefixMatch()
	componentIdent := entry.makeComponentIdent(parentComponentIdent)
	entry.Ident = componentIdent
	if ("" != entry.StrictPrefixMatch) && entry.StrictMatch {
		return &ErrConflictConfiguration{
			Component: componentIdent,
			Config1:   "strict-prefix-match=" + entry.StrictPrefixMatch,
			Config2:   "strict-match=true",
			Message:   "partial-strict-match and fully-strict-match cannot co-exist",
		}
	}
	if "" == entry.HandlerName {
		if entry.TrailingSlash {
			return &ErrConflictConfiguration{
				Component: componentIdent,
				Config1:   "trailing-slash=true",
				Config2:   "handler=" + entry.HandlerName,
				Message:   "enabling trailing-slash on terminate component only",
			}
		}
		if 0 == len(entry.Routes) {
			return &ErrConflictConfiguration{
				Component: componentIdent,
				Config1:   "terminate-component=true",
				Config2:   "handler=" + entry.HandlerName,
				Message:   "require handler at terminate component",
			}
		}
	}
	for _, childEntry := range entry.Routes {
		if err := childEntry.verifyConfiguration(componentIdent); nil != err {
			return err
		}
	}
	return nil
}

// LoadYAML get route configuration from YAML file
func LoadYAML(configFilePath string) (routeEntry *RouteEntry, err error) {
	buf, err := ioutil.ReadFile(configFilePath)
	if nil != err {
		return
	}
	var routeEntryBuf RouteEntry
	if err = yaml.Unmarshal(buf, &routeEntryBuf); nil != err {
		return
	}
	if err = routeEntryBuf.verifyConfiguration(""); nil != err {
		return
	}
	return &routeEntryBuf, nil
}
