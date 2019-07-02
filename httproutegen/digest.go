package httproutegen

import (
	"log"
)

// ComputeLiteralDigest generate literal digest from string.
func ComputeLiteralDigest(literal string) (digest uint64) {
	b := []byte(literal)
	if len(b) > 8 {
		log.Printf("WARN: compute literal digest against string with length > 8: %v", literal)
	}
	for _, ch := range b {
		digest = (digest << 8) | uint64(ch)
	}
	return
}
