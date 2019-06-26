package httproutegen

import (
	"log"
)

type byteMap struct {
	bits [2]uint64
}

func (m *byteMap) enableByte(b byte) {
	if b > 127 {
		log.Printf("WARN: register byte > 127: %v", b)
		return
	}
	bVal := uint(b)
	bIndex := 0
	if 0 != (bVal & 64) {
		bIndex = 1
	}
	offset := bVal & 63
	m.bits[bIndex] = m.bits[bIndex] | (1 << offset)
}

func (m *byteMap) enableByteRange(b0, b1 byte) {
	if (b0 > 127) || (b1 > 127) {
		log.Printf("WARN: register byte range > 127: %v - %v", b0, b1)
		return
	}
	if b1 < b0 {
		b0, b1 = b1, b0
	}
	for b := b0; b <= b1; b++ {
		m.enableByte(b)
	}
}

func (m *byteMap) enablePrintables() {
	m.enableByteRange(0x20, 0x7E)
}

func (m *byteMap) disableByte(b byte) {
	if b > 127 {
		log.Printf("WARN: register byte > 127: %v", b)
		return
	}
	bVal := uint(b)
	bIndex := 0
	if 0 != (bVal & 64) {
		bIndex = 1
	}
	offset := bVal & 63
	m.bits[bIndex] = m.bits[bIndex] & (^uint64(1 << offset))
}

func (m *byteMap) disableByteRange(b0, b1 byte) {
	if (b0 > 127) || (b1 > 127) {
		log.Printf("WARN: register byte range > 127: %v - %v", b0, b1)
		return
	}
	if b1 < b0 {
		b0, b1 = b1, b0
	}
	for b := b0; b <= b1; b++ {
		m.disableByte(b)
	}
}
