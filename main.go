package main

import (
	"log"
)

func main() {
	inputFilePath, outputFilePath, err := parseCommandParam()
	if nil != err {
		log.Fatalf("ERR: cannot have required parameters: %v", err)
		return
	}
	log.Printf("Input: %v", inputFilePath)
	log.Printf("Output: %v", outputFilePath)
}
