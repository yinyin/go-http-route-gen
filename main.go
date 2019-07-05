package main

import (
	"encoding/json"
	"log"

	"github.com/yinyin/go-http-route-gen/httproutegen"
)

func main() {
	inputFilePath, outputFilePath, packageName, receiverName, handlerTypeName, routeMethodName, genNamePrefix, dumpFanoutContent, err := parseCommandParam()
	if nil != err {
		log.Fatalf("ERR: cannot have required parameters: %v", err)
		return
	}
	log.Printf("Input: [%v].", inputFilePath)
	log.Printf("Output: [%v]", outputFilePath)
	log.Printf("Route Method: (%s *%s) %s() (%sRouteIdent).", receiverName, handlerTypeName, routeMethodName, genNamePrefix)
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
	} else if dumpFanoutContent {
		log.Print(string(fanoutJSONText))
	}
	codeGenInst, err := httproutegen.OpenCodeGenerateInstance(outputFilePath, fanoutInstance.RootFanoutFork)
	if nil != err {
		log.Fatalf("ERR: cannot open code generation instance: %v", err)
		return
	}
	defer codeGenInst.Close()
	codeGenInst.PackageName = packageName
	codeGenInst.ReceiverName = receiverName
	codeGenInst.HandlerTypeName = handlerTypeName
	codeGenInst.RouteMethodName = routeMethodName
	codeGenInst.NamePrefix = genNamePrefix
	err = codeGenInst.Generate()
	log.Printf("Code generate stopped: %v", err)
}
