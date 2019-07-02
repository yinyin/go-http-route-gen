package main

import (
	"os"
	"log"

	"github.com/yinyin/go-http-route-gen/httproutegen"
)

func main() {
	if len(os.Args) < 2 {
		log.Print("Argument: [TargetString] ...")
		return
	}
	for _, arg := range os.Args[1:] {
		digest := httproutegen.ComputeLiteralDigest(arg)
		log.Printf("LiteralDigest: %s, Digest: 0x%016X.", arg, digest)
	}
}
