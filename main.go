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
	fanoutInstance, err := httproutegen.MakeFanoutInstance(rootRouteEntry)
	if nil != err {
		log.Fatalf("ERR: cannot create fanout instance from root route entry: %v", err)
		return
	}
	if err = fanoutInstance.ExpandFanout(); nil != err {
		log.Fatalf("ERR: cannot expand fanout instance: %v", err)
		return
	}
	if fanoutJSONText, err := json.MarshalIndent(fanoutInstance, "", "  "); nil != err {
		log.Fatalf("ERR: cannot encode root fanout into JSON: %v", err)
	} else if ':' == outputFilePath[0] {
		log.Printf("Starting HTTP at %v", outputFilePath)
		err = runHTTPService(outputFilePath, fanoutJSONText)
		log.Printf("HTTP stopped: %v", err)
		return
	} else {
		log.Print(string(fanoutJSONText))
	}
}
