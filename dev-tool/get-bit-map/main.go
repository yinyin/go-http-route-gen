package main

import (
	"os"
	"log"

	"github.com/yinyin/go-http-route-gen/httproutegen"
)

func main() {
	if len(os.Args) < 2 {
		log.Print("Argument: [ByteMapRule] ...")
		return
	}
	for _, arg := range os.Args[1:] {
		var mapper httproutegen.ByteMapper
		mapper.SetByteMap([]byte(arg), 0)
		b0, b1 := mapper.ByteMap()
		log.Printf("Rule: %s, Map: 0x%08X, 0x%08X", arg, b0, b1)
	}
}
