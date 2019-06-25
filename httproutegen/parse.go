package httproutegen

import (
	"errors"
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// RouteEntry represent an entry of route
type RouteEntry struct {
	Component         string        `yaml:"c,omitempty"`
	HandlerName       string        `yaml:"name,omitempty"`
	StrictPrefixMatch string        `yaml:"strict-prefix-match,omitempty"`
	StrictMatch       bool          `yaml:"strict-match,omitempty"`
	Routes            []*RouteEntry `yaml:"route,omitempty"`
}

func (entry *RouteEntry) cleanupComponent(parentComponents string) error {
	entry.Component = strings.Trim(entry.Component, "/")
	if "" == entry.Component {
		return errors.New("empty component: parent=[" + parentComponents + "], handler=[" + entry.HandlerName + "]")
	}
	return nil
}

func (entry *RouteEntry) verifyConfiguration(parentComponents string) error {
	if err := entry.cleanupComponent(parentComponents); nil != err {
		return err
	}
	componentIdent := parentComponents + entry.Component
	if ("" != entry.StrictPrefixMatch) && entry.StrictMatch {
		return &ErrConflictConfiguration{
			Component: componentIdent,
			Config1:   "strict-prefix-match=" + entry.StrictPrefixMatch,
			Config2:   "strict-match=true",
			Message:   "partial-strict-match and fully-strict-match cannot co-exist",
		}
	}
	for _, childEntry := range entry.Routes {
		if err := childEntry.verifyConfiguration(componentIdent + "/"); nil != err {
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
	if err = routeEntryBuf.verifyConfiguration("/"); nil != err {
		return
	}
	return &routeEntryBuf, nil
}
