package httproutegen

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// RouteEntry represent an entry of route
type RouteEntry struct {
	Component         string        `yaml:"c,omitempty"`
	HandlerName       string        `yaml:"name,omitempty"`
	StrictPrefixMatch string        `yaml:"strict-prefix-match,omitempty"`
	Routes            []*RouteEntry `yaml:"route,omitempty"`
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
	// TODO: post processing
	return &routeEntryBuf, nil
}
