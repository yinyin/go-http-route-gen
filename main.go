package main

import (
	"encoding/json"
	"log"

	"github.com/yinyin/go-http-route-gen/httproutegen"
)

func main() {
	inputFilePath, outputFilePath, err := parseCommandParam()
	if nil != err {
		log.Fatalf("ERR: cannot have required parameters: %v", err)
		return
	}
	log.Printf("Input: %v", inputFilePath)
	log.Printf("Output: %v", outputFilePath)
	rootRouteEntry, err := httproutegen.LoadYAML(inputFilePath)
	if nil != err {
		log.Fatalf("ERR: cannot load route configuration [%s]: %v", inputFilePath, err)
		return
	}
	rootFanoutEntry := httproutegen.NewFanoutEntry(rootRouteEntry)
	if fanoutJSONText, err := json.MarshalIndent(rootFanoutEntry, "", "  "); nil != err {
		log.Fatalf("ERR: cannot encode root fanout into JSON: %v", err)
	} else {
		log.Print(string(fanoutJSONText))
	}
}
